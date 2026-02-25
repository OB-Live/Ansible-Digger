package adigger

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

// Parse unmarshals the playbook YAML data into a structured Go model.
// It uses a single, robust unmarshal operation thanks to a well-defined model.
func (p *Parser) Parse(data []byte) (*Playbook, error) {
	playbook := &Playbook{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	// decoder.KnownFields(true) // This is too strict for Ansible modules. The 'inline' field handles them.

	if err := decoder.Decode(&playbook.Plays); err != nil {
		return nil, fmt.Errorf("failed to unmarshal playbook: %w", err)
	}

	p.postProcess(playbook)

	return playbook, nil
}

// postProcess walks the parsed structure to assign IDs and perform any necessary linking.
func (p *Parser) postProcess(playbook *Playbook) {
	for playIdx, play := range playbook.Plays {
		play.ID = playIdx

		// Assign unique IDs to roles within the play
		for roleIdx, role := range play.Roles {
			role.ID = roleIdx
			role.PlayID = playIdx
		}

		// Helper function to recursively assign IDs within postProcess
		var assignIDs func(task *Task, prefix string, playID int, taskIdx *int)
		assignIDs = func(task *Task, prefix string, playID int, taskIdx *int) {
			task.ID = fmt.Sprintf("%s_%d_%d", prefix, playID, *taskIdx)
			task.PlayID = playID
			*taskIdx++

			for _, t := range task.Block {
				assignIDs(t, prefix, playID, taskIdx)
			}
			for _, t := range task.Rescue {
				assignIDs(t, prefix, playID, taskIdx)
			}
			for _, t := range task.Always {
				assignIDs(t, prefix, playID, taskIdx)
			}
		}

		taskCounter := 0
		for _, task := range play.PreTasks {
			assignIDs(task, "pre_task", playIdx, &taskCounter)
		}
		for _, task := range play.Tasks {
			assignIDs(task, "task", playIdx, &taskCounter)
		}
		for _, task := range play.PostTasks {
			assignIDs(task, "post_task", playIdx, &taskCounter)
		}

		// Process tasks within roles
		for _, role := range play.Roles {
			for _, task := range role.Tasks {
				assignIDs(task, fmt.Sprintf("role_%d_task", role.ID), playIdx, &taskCounter)
			}
		}

		// Now process handlers and set their flag
		for _, task := range play.Handlers {
			task.IsHandler = true
			assignIDs(task, "handler", playIdx, &taskCounter)
		}
		infof("Processed Play '%s' with %d tasks, %d roles, and %d handlers.", play.Name, len(play.Tasks), len(play.Roles), len(play.Handlers))
	}
}
