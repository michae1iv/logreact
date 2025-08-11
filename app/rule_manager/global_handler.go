package rule_manager

import (
	"context"
	"correlator/db"
	"correlator/logger"
	"encoding/json"
	"fmt"
	"time"
)

type ChTypes interface {
	string | map[string]interface{} | *Rule | []byte
}

// Function to read all messages from channel effective in for:select loop
func drainChannel[c ChTypes](ch <-chan c) []c {
	var drained []c
	for {
		select {
		case msg := <-ch:
			drained = append(drained, msg)
		default:
			return drained
		}
	}
}

type GlobalHandler struct {
	IsReady      bool
	FramePointer map[string]*Frame
	RuleChan     chan *Rule  // Channel for calls, for example creating new rule or stopping one
	LogChan      chan []byte // Channel for Reader
}

var GlobalHandlerObj = GlobalHandler{}

// Main function of global handler
func (g *GlobalHandler) Start(ctx context.Context) {
	defer func() {
		close(g.LogChan) // closing channel if it didn't closed already
	}()

	f_ctx, close := context.WithCancel(ctx) // Child context for frames, if paarent context cancells also cansells childs
	_ = close
	if err := g.initFrames(f_ctx); err != nil {
		logger.ErrorLogger.Printf("Error occure while initing frames: %s", err.Error())
	}
	logger.InfoLogger.Println("GlobalHandler started")
	g.IsReady = true // GlobalHandler now can recieve logs

	for {
		select {
		case <-ctx.Done():
			logger.InfoLogger.Println("Stop signal recieved, stopping handler")
			return
		case rule, ok := <-g.RuleChan: // because rule channel is rarely used it has the biggest priority
			if !ok {
				logger.InfoLogger.Println("Meassage channel closed, stopping handler.")
				return
			}
			g.manageRule(ctx, rule)
		case msg, ok := <-g.LogChan: // Checking if new log
			if !ok {
				logger.InfoLogger.Println("Log channel closed, stopping handler.")
				return
			}
			g.parseLog(msg)
		default:
			time.Sleep(1 * time.Nanosecond)
		}
	}
}

// Function thats creates frames based on unique key, rules are getting from db if no rules found returns error
func (g *GlobalHandler) initFrames(ctx context.Context) error {
	// Default frame for rules with undefined Ukey
	f := g.createFrame(ctx, "")
	g.FramePointer[""] = f

	// Pull Rules from db
	var rules []db.Rules
	if res := db.DB.Find(&rules); res.RowsAffected < 1 {
		err := fmt.Errorf("rules not found")
		logger.ErrorLogger.Printf("Error getting rules from db: %s", err.Error())
		return err
	}
	// Unmarshall record from db of type []byte to struct Rule{}
	for _, rule := range rules {
		r := Rule{}
		r.ParseRule(rule.Rule)
		// looking on ukey value of every rule, if it's not in map creating new frame and pushing to map
		if g.FramePointer[r.UKeyValue] == nil {
			f := g.createFrame(ctx, r.UKeyValue)
			g.FramePointer[r.UKeyValue] = f
		}
		g.FramePointer[r.UKeyValue].RuleChan <- &r // sending rule to frame with channel
	}
	return nil
}

// Function thats manage new or deleted rules sending them to frame
func (g *GlobalHandler) manageRule(ctx context.Context, r *Rule) error {
	// If Rule comes with new ukey creating new frame for it
	if g.FramePointer[r.UKeyValue] == nil && r.UKeyValue != "" {
		f := g.createFrame(ctx, r.UKeyValue)
		g.FramePointer[r.UKeyValue] = f
	} else if g.FramePointer[r.UKeyValue] != nil {
		g.FramePointer[r.UKeyValue].RuleChan <- r
	}

	return nil
}

// Unmarshalls stored logs into json
func (g GlobalHandler) parseLog(log []byte) {
	var json_log map[string]interface{}
	err := json.Unmarshal(log, &json_log)
	if err != nil {
		logger.ErrorLogger.Printf("Error parsing log: %s\n", err.Error())
		return
	}

	if value, ok := json_log["ukey"].(string); ok && g.FramePointer[value] != nil {
		g.FramePointer[value].LogChan <- json_log
	}
}

// Creating new frame with channel initialization
func (g GlobalHandler) createFrame(ctx context.Context, ukey string) *Frame {
	fr := Frame{
		Ukeys:          ukey,
		WorkersPointer: make(map[string]*Worker),
		LogChan:        make(chan map[string]interface{}, 1000),
		RuleChan:       make(chan *Rule),
	}
	go fr.StartFrame(ctx)
	return &fr
}
