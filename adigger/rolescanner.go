package adigger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath" 

	"gopkg.in/yaml.v3"
)

type RoleScanner struct {
	rolesPaths []string
}

func NewRoleScanner(extraPath string) *RoleScanner {
	// Per spec, search in standard locations
	paths := []string{
		extraPath,
		"./roles",
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".ansible", "roles"))
	}
	paths = append(paths, "~/ansible/roles")
	return &RoleScanner{rolesPaths: paths}
}

func (rs *RoleScanner) ScanAndEnrich(playbook *Playbook) error {
	infof("Scanning for roles...")
	for _, play := range playbook.Plays {
		for _, role := range play.Roles {

			if err := rs.findAndParseRole(role, play.ID); err != nil {
				warnf("Could not process role '%s': %v", role.Name, err)
			}
		}
	}
	return nil
}

func (rs *RoleScanner) findAndParseRole(role *Role, playID int) error {
	infof("searching for role %s ", role.Name)
	for _, path := range rs.rolesPaths {
		rolePath := filepath.Join(path, role.Name)
		if _, err := os.Stat(rolePath); !os.IsNotExist(err) {
			infof("Found role '%s' at %s", role.Name, rolePath)
			tasksPath := filepath.Join(rolePath, "tasks", "main.yml")
			if _, err := os.Stat(tasksPath); !os.IsNotExist(err) {
				return parseRoleTasks(tasksPath, role, playID)
			}
			// Also check for tasks/main.yaml
			tasksPath = filepath.Join(rolePath, "tasks", "main.yaml")
			if _, err := os.Stat(tasksPath); !os.IsNotExist(err) {
				return parseRoleTasks(tasksPath, role, playID)
			}
		}
	}
	return nil // Not an error if a role has no tasks file
}

func parseRoleTasks(filePath string, role *Role, playID int) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var tasks []*Task
	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return err
	}

	// Assign the PlayID to each task. The unique task ID will be assigned in the parser's post-processing step.
	for _, task := range tasks {
		task.PlayID = playID
		infof("Found role's task %v", task)

		if task.Include != "" {
			x := CleanID(task.Include)
			task.ID = fmt.Sprintf("role_%d_task_%s", role.ID, x)

		} else if task.IncludeTasks != "" {
			x := CleanID(task.IncludeTasks)
			task.ID = fmt.Sprintf("role_%d_task_%s", role.ID, x)

		} else if task.ImportTasks != "" {
			x := CleanID(task.ImportTasks)
			task.ID = fmt.Sprintf("role_%d_task_%s", role.ID, x)

		}else if task.Name != "" {
			x := CleanID(task.Name)
			task.ID = fmt.Sprintf("role_%d_task_%s", role.ID, x)
		}else{
			warnf("unhandled task type for ")
			PrintJSON(task)
		}
		
	}
	role.Tasks = tasks
	infof("Parsed %d tasks for role '%s'", len(tasks), role.Name) 
	return nil
}
