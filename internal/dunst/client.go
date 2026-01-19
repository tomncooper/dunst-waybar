package dunst

import (
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	dunstService   = "org.freedesktop.Notifications"
	dunstPath      = "/org/freedesktop/Notifications"
	dunstInterface = "org.dunstproject.cmd0"
	propsInterface = "org.freedesktop.DBus.Properties"
)

// StateUpdate represents a change in dunst state
type StateUpdate struct {
	Paused        bool
	WaitingLength int
	IsError       bool
	Error         error
}

// Client manages the D-Bus connection to dunst
type Client struct {
	conn          *dbus.Conn
	updates       chan StateUpdate
	mu            sync.RWMutex
	paused        bool
	waitingLength int
	connected     bool
}

// NewClient creates a new dunst D-Bus client
func NewClient() (*Client, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	client := &Client{
		conn:      conn,
		updates:   make(chan StateUpdate, 10),
		connected: false,
	}

	return client, nil
}

// Start begins listening for dunst property changes
func (c *Client) Start() error {
	// Query initial state
	if err := c.queryInitialState(); err != nil {
		c.sendUpdate(StateUpdate{IsError: true, Error: err})
		return err
	}

	c.connected = true

	// Subscribe to property changes
	if err := c.subscribeToChanges(); err != nil {
		c.connected = false
		c.sendUpdate(StateUpdate{IsError: true, Error: err})
		return err
	}

	// Send initial state
	c.mu.RLock()
	c.sendUpdate(StateUpdate{
		Paused:        c.paused,
		WaitingLength: c.waitingLength,
		IsError:       false,
	})
	c.mu.RUnlock()

	return nil
}

// Updates returns the channel for receiving state updates
func (c *Client) Updates() <-chan StateUpdate {
	return c.updates
}

// Close closes the D-Bus connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// queryInitialState queries the current dunst state
func (c *Client) queryInitialState() error {
	obj := c.conn.Object(dunstService, dunstPath)

	// Get paused property
	pausedVariant, err := obj.GetProperty(dunstInterface + ".paused")
	if err != nil {
		return fmt.Errorf("failed to get paused property: %w", err)
	}

	paused, ok := pausedVariant.Value().(bool)
	if !ok {
		return fmt.Errorf("paused property is not a boolean")
	}

	// Get waitingLength property
	waitingVariant, err := obj.GetProperty(dunstInterface + ".waitingLength")
	if err != nil {
		return fmt.Errorf("failed to get waitingLength property: %w", err)
	}

	waitingLength, ok := waitingVariant.Value().(uint32)
	if !ok {
		return fmt.Errorf("waitingLength property is not a uint32")
	}

	c.mu.Lock()
	c.paused = paused
	c.waitingLength = int(waitingLength)
	c.mu.Unlock()

	return nil
}

// subscribeToChanges subscribes to property change signals
func (c *Client) subscribeToChanges() error {
	// Add match rule for property changes
	matchRule := fmt.Sprintf(
		"type='signal',interface='%s',member='PropertiesChanged',path='%s'",
		propsInterface,
		dunstPath,
	)

	if err := c.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule).Err; err != nil {
		return fmt.Errorf("failed to add match rule: %w", err)
	}

	// Create signal channel
	signals := make(chan *dbus.Signal, 10)
	c.conn.Signal(signals)

	// Start listening goroutine
	go c.handleSignals(signals)

	return nil
}

// handleSignals processes incoming D-Bus signals
func (c *Client) handleSignals(signals chan *dbus.Signal) {
	for sig := range signals {
		if sig.Name != propsInterface+".PropertiesChanged" {
			continue
		}

		if len(sig.Body) < 2 {
			continue
		}

		// First argument is the interface name
		iface, ok := sig.Body[0].(string)
		if !ok || iface != dunstInterface {
			continue
		}

		// Second argument is a map of changed properties
		changed, ok := sig.Body[1].(map[string]dbus.Variant)
		if !ok {
			continue
		}

		hasChanges := false

		c.mu.Lock()
		if pausedVar, ok := changed["paused"]; ok {
			if paused, ok := pausedVar.Value().(bool); ok {
				c.paused = paused
				hasChanges = true
			}
		}

		if waitingVar, ok := changed["waitingLength"]; ok {
			if waiting, ok := waitingVar.Value().(uint32); ok {
				c.waitingLength = int(waiting)
				hasChanges = true
			}
		}

		if hasChanges {
			update := StateUpdate{
				Paused:        c.paused,
				WaitingLength: c.waitingLength,
				IsError:       false,
			}
			c.mu.Unlock()
			c.sendUpdate(update)
		} else {
			c.mu.Unlock()
		}
	}
}

// sendUpdate sends a state update to the updates channel
func (c *Client) sendUpdate(update StateUpdate) {
	select {
	case c.updates <- update:
	default:
		// Channel full, skip update
	}
}
