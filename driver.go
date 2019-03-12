package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	api "github.com/hetznercloud/hcloud-go/hcloud"
)

const (
	defaultImage      = "debian-9"
	defaultLocation   = "fsn1"
	defaultServerType = "cx11"
)

// Driver ...
type Driver struct {
	*drivers.BaseDriver

	AccessToken string
	Image       string
	Location    string
	ServerID    int
	ServerType  string
	SSHKeyID    int
}

// NewDriver ...
func NewDriver() *Driver {
	return &Driver{
		Image:      defaultImage,
		Location:   defaultLocation,
		ServerType: defaultServerType,

		BaseDriver: &drivers.BaseDriver{
			SSHUser: drivers.DefaultSSHUser,
			SSHPort: drivers.DefaultSSHPort,
		},
	}
}

func (d *Driver) getClient() *api.Client {
	return api.NewClient(api.WithToken(d.AccessToken))
}

func (d *Driver) getImage() (*api.Image, error) {
	image, _, err := d.getClient().Image.GetByName(context.Background(), d.Image)

	return image, err
}

func (d *Driver) getLocation() (*api.Location, error) {
	location, _, err := d.getClient().Location.GetByName(context.Background(), d.Location)

	return location, err
}

func (d *Driver) getServer() (*api.Server, error) {
	server, _, err := d.getClient().Server.GetByID(context.Background(), d.ServerID)

	return server, err
}

func (d *Driver) getServerType() (*api.ServerType, error) {
	serverType, _, err := d.getClient().ServerType.GetByName(context.Background(), d.ServerType)

	return serverType, err
}

func (d *Driver) getSSHKey() (*api.SSHKey, error) {
	sshKey, _, err := d.getClient().SSHKey.GetByID(context.Background(), d.SSHKeyID)

	return sshKey, err
}

func (d *Driver) createSSHKey() error {
	// Generate new SSH Key pair
	err := ssh.GenerateSSHKey(d.GetSSHKeyPath())
	if err != nil {
		return err
	}

	// Read public SSH Key
	publicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
	if err != nil {
		return err
	}

	opts := api.SSHKeyCreateOpts{
		Name:      d.GetMachineName(),
		PublicKey: string(publicKey),
	}

	// Upload public SSH Key to Hetzner
	key, _, err := d.getClient().SSHKey.Create(context.Background(), opts)

	d.SSHKeyID = key.ID

	return err
}

func (d *Driver) createServer() error {
	sshKey, err := d.getSSHKey()
	if err != nil {
		return err
	}

	image, err := d.getImage()
	if err != nil {
		return err
	}

	location, err := d.getLocation()
	if err != nil {
		return err
	}

	serverType, err := d.getServerType()
	if err != nil {
		return err
	}

	opts := api.ServerCreateOpts{
		Image:      image,
		Location:   location,
		Name:       d.GetMachineName(),
		ServerType: serverType,
	}
	opts.SSHKeys = append(opts.SSHKeys, sshKey)

	server, _, err := d.getClient().Server.Create(context.Background(), opts)
	if err != nil {
		return err
	}

	d.ServerID = server.Server.ID

	for {
		serverState, err := d.GetState()
		if err != nil {
			return err
		}

		if serverState == state.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	d.IPAddress = server.Server.PublicNet.IPv4.IP.String()

	return nil
}

func (d *Driver) waitForAction(a *api.Action) error {
	for {
		action, _, err := d.getClient().Action.GetByID(context.Background(), a.ID)
		if err != nil {
			return err
		}

		switch action.Status {
		case api.ActionStatusSuccess:
			return nil
		case api.ActionStatusError:
			return action.Error()
		}

		time.Sleep(1 * time.Second)
	}
}

// DriverName ...
func (d *Driver) DriverName() string {
	return "hetzner"
}

// GetCreateFlags ...
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "HETZNER_ACCESS_TOKEN",
			Name:   "hetzner-access-token",
			Usage:  "Access token",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_IMAGE",
			Name:   "hetzner-image",
			Usage:  "Image",
			Value:  defaultImage,
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_LOCATION",
			Name:   "hetzner-location",
			Usage:  "Location",
			Value:  defaultLocation,
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_SERVER_TYPE",
			Name:   "hetzner-server-type",
			Usage:  "Server type",
			Value:  defaultServerType,
		},
	}
}

// SetConfigFromFlags ...
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessToken = flags.String("hetzner-access-token")
	d.Image = flags.String("hetzner-image")
	d.Location = flags.String("hetzner-location")
	d.ServerType = flags.String("hetzner-server-type")

	d.SetSwarmConfigFromFlags(flags)

	if d.AccessToken == "" {
		return fmt.Errorf("hetzner driver requres the --hetzner-access-token option")
	}

	return nil
}

// GetSSHHostname ...
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetURL ...
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

// GetState ...
func (d *Driver) GetState() (state.State, error) {
	server, err := d.getServer()
	if err != nil {
		return state.Error, err
	}

	switch server.Status {
	case api.ServerStatusInitializing:
		return state.Starting, nil
	case api.ServerStatusRunning:
		return state.Running, nil
	case api.ServerStatusOff:
		return state.Stopped, nil
	}

	return state.None, nil
}

// PreCreateCheck ...
func (d *Driver) PreCreateCheck() error {
	if d.getClient() == nil {
		return fmt.Errorf("cannot create client")
	}

	return nil
}

// Create Server
func (d *Driver) Create() error {
	// Create SSH key
	if err := d.createSSHKey(); err != nil {
		return err
	}

	// Create Server
	if err := d.createServer(); err != nil {
		return err
	}

	return nil
}

// Start Server
func (d *Driver) Start() error {
	server, err := d.getServer()
	if err != nil {
		return err
	}

	action, _, err := d.getClient().Server.Poweron(context.Background(), server)
	if err != nil {
		return err
	}

	return d.waitForAction(action)
}

// Stop Server
func (d *Driver) Stop() error {
	server, err := d.getServer()
	if err != nil {
		return err
	}

	action, _, err := d.getClient().Server.Poweroff(context.Background(), server)
	if err != nil {
		return err
	}

	return d.waitForAction(action)
}

// Restart Server
func (d *Driver) Restart() error {
	server, err := d.getServer()
	if err != nil {
		return err
	}

	action, _, err := d.getClient().Server.Reboot(context.Background(), server)
	if err != nil {
		return err
	}

	return d.waitForAction(action)
}

// Remove Server
func (d *Driver) Remove() error {
	server, err := d.getServer()
	if err != nil {
		return err
	}

	_, err = d.getClient().Server.Delete(context.Background(), server)

	return err
}

// Kill Server
func (d *Driver) Kill() error {
	server, err := d.getServer()
	if err != nil {
		return err
	}

	action, _, err := d.getClient().Server.Shutdown(context.Background(), server)
	if err != nil {
		return err
	}

	return d.waitForAction(action)
}
