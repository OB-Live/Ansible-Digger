package adigger

import (
	"fmt"
	"strings"
)

func Render(playbook *Playbook) string {
	var sb strings.Builder

	graph(&sb, 0, "digraph \"Ansible Digger\" {")
	graph(&sb, 1, "rankdir=TB;")
	graph(&sb, 1, "compound=true;")
	graph(&sb, 1, "fontname=\"Helvetica\";")
	graph(&sb, 1, "node [fontname=\"Helvetica\"];")
	graph(&sb, 1, "edge [fontname=\"Helvetica\"];")
	graph(&sb, 1, "graph [label=\"Ansible Digger\", labelloc=t, fontsize=20];\n")

	for _, play := range playbook.Plays {
		renderPlay(&sb, play)
	}

	for _, dep := range playbook.Dependencies {
		renderDependency(&sb, &dep)
	}

	graph(&sb, 0, "}")

	return sb.String()
}

func renderPlay(sb *strings.Builder, play *Play) {
	playID := fmt.Sprintf("cluster_play_%d", play.ID)
	graph(sb, 1, fmt.Sprintf("subgraph %s {", playID))
	graph(sb, 2, fmt.Sprintf("id = \"%s\";", playID))
	graph(sb, 2, "class = \"play\";")
	graph(sb, 2, fmt.Sprintf("URL = \"#play-%d\";", play.ID))
	graph(sb, 2, fmt.Sprintf("label = \"🎭 Play: %s\\nHosts: %s\";", play.Name, play.Hosts))
	graph(sb, 2, "style = filled;")
	graph(sb, 2, "fillcolor = \"#E8DAEF\";\n")

	// Render all the node and cluster definitions first
	renderTaskCluster(sb, play.PreTasks, "pre_tasks", "🔧 Pre-Tasks", play.ID)
	if len(play.Roles) > 0 {
		graph(sb, 2, fmt.Sprintf("subgraph cluster_play_%d_roles {", play.ID))
		graph(sb, 3, "label = \"📦 Roles\";")
		for _, role := range play.Roles {
			renderRole(sb, role, play.ID)
		}
		graph(sb, 2, "}")
	}
	renderTaskCluster(sb, play.Tasks, "tasks", "⚙ Tasks", play.ID)
	renderTaskCluster(sb, play.PostTasks, "post_tasks", "⚙ Post-Tasks", play.ID)

	// Render Handlers as individual nodes, not in a cluster
	if len(play.Handlers) > 0 {
		graph(sb, 2, "// Handler Nodes")
		for _, handler := range play.Handlers {
			renderTask(sb, handler, 3)
		}
		sb.WriteString("\n") // Keep a blank line for readability
	}

	// Now, draw the sequential edges to show execution flow
	graph(sb, 2, "\n// Sequential Edges")
	var lastNodeID string
	lastNodeID = renderSequentialEdges(sb, play.PreTasks, lastNodeID)
	lastNodeID = renderSequentialRoleEdges(sb, play.Roles, lastNodeID)
	lastNodeID = renderSequentialEdges(sb, play.Tasks, lastNodeID)
	renderSequentialEdges(sb, play.PostTasks, lastNodeID)

	// Render notification edges
	renderNotifyEdges(sb, play)

	graph(sb, 1, "}\n")
}

func renderSequentialEdges(sb *strings.Builder, tasks []*Task, predecessorID string) string {
	for _, task := range tasks {
		currentNodeID := task.ID
		connectNodes(sb, predecessorID, currentNodeID, "")
		predecessorID = currentNodeID
	}
	return predecessorID
}

func renderTaskCluster(sb *strings.Builder, tasks []*Task, clusterName, clusterLabel string, playID int) {
	if len(tasks) == 0 {
		return
	}
	graph(sb, 2, fmt.Sprintf("subgraph cluster_play_%d_%s {", playID, clusterName))
	graph(sb, 3, fmt.Sprintf("label = \"%s\";", clusterLabel))
	for _, task := range tasks {
		renderTask(sb, task, 3)
	}
	graph(sb, 2, "}")
}

func renderTask(sb *strings.Builder, task *Task, indent int) {
	emoji := "⚙"
	if task.IsHandler {
		emoji = "🚨"
	}
	taskNodeID := task.ID
	var label, module string

	if task.Include != "" {
		emoji = "🔗" // Link emoji for includes
		label = fmt.Sprintf("%s %s", emoji, task.Include)
	} else if task.IncludeTasks != "" {
		emoji = "🔗" // Link emoji for includes
		label = fmt.Sprintf("%s %s", emoji, task.IncludeTasks)
	} else if task.ImportTasks != "" {
		emoji = "🔗" // Link emoji for includes
		label = fmt.Sprintf("%s %s", emoji, task.ImportTasks)
	} else {
		label = fmt.Sprintf("%s %s", emoji, task.Name)
		module = task.GetModule()
		label += fmt.Sprintf("\\n%s", module)
	}
	
	if task.Become {
		label += " 🔒"
	}

	graph(sb, indent, fmt.Sprintf("%s [", taskNodeID))
	graph(sb, indent+1, fmt.Sprintf("id = \"%s\",", taskNodeID))
	// graph(sb, 4, fmt.Sprintf("class = \"task module-%s\",", strings.ReplaceAll(module, ".", "-")))
	graph(sb, indent+1, fmt.Sprintf("label = \"%s\",", label))
	graph(sb, indent+1, "shape = \"record\",")
	graph(sb, indent+1, "style = \"filled\",")
	graph(sb, indent+1, "fillcolor = \"white\",")
	graph(sb, indent+1, fmt.Sprintf("tooltip = \"Tooltip for %s\",", task.Name)) // Placeholder
	graph(sb, indent+1, fmt.Sprintf("URL = \"#%s\"", task.ID))
	graph(sb, indent, "];")
}

func renderRole(sb *strings.Builder, role *Role, playID int) {
	infof("Rendering role: %d - %s", playID, string(role.Name))
	roleClusterID := fmt.Sprintf("cluster_play_%d_role_%d", playID, role.ID) // Unique cluster ID for role
	graph(sb, 3, fmt.Sprintf("subgraph %s {", roleClusterID))
	graph(sb, 4, fmt.Sprintf("id = \"%s\";", roleClusterID))
	graph(sb, 4, "class = \"role\";")
	graph(sb, 4, fmt.Sprintf("label = \"📦 %s\";", role.Name))
	graph(sb, 4, "style = filled;")
	graph(sb, 4, "fillcolor = \"#D6EAF8\";")

	infof("len(role.Tasks) %d", len(role.Tasks))
	fmt.Printf("%+v\n", role.Tasks)
	if len(role.Tasks) > 0 {
		// Render tasks inside a nested cluster
		graph(sb, 4, fmt.Sprintf("subgraph cluster_play_%d_role_%d_tasks {", playID, role.ID))
		graph(sb, 5, "label = \"\";")
		// graph(sb, 5, "style = \"invis\";")
		for _, task := range role.Tasks {
			renderTask(sb, task, 5)
		}
		graph(sb, 4, "}")
	} else {
		graph(sb, 4, fmt.Sprintf("role_placeholder_%d_%d [label=\"(details)\", shape=plaintext];", playID, role.ID))
	}

	graph(sb, 3, "}")
}

func graph(sb *strings.Builder, indent int, label string) {
	sb.WriteString(strings.Repeat("  ", indent))
	sb.WriteString(label)
	sb.WriteString("\n")
}

func renderSequentialRoleEdges(sb *strings.Builder, roles []*Role, predecessorID string) string {
	for _, role := range roles {
		if len(role.Tasks) > 0 {
			predecessorID = renderSequentialEdges(sb, role.Tasks, predecessorID)
		} else {
			currentNodeID := fmt.Sprintf("role_placeholder_%d_%d", role.PlayID, role.ID)
			connectNodes(sb, predecessorID, currentNodeID, "")
			predecessorID = currentNodeID
		}
	}
	return predecessorID
}

func renderNotifyEdges(sb *strings.Builder, play *Play) {
	if len(play.Handlers) == 0 {
		return
	}
	graph(sb, 2, "\n// Notification Edges")

	handlerMap := make(map[string]*Task)
	for _, handler := range play.Handlers {
		handlerMap[handler.Name] = handler
	}

	allTasks := [][]*Task{play.PreTasks, play.Tasks, play.PostTasks}
	for _, role := range play.Roles {
		allTasks = append(allTasks, role.Tasks)
	}

	for _, taskList := range allTasks {
		for _, task := range taskList {
			if len(task.Notify) > 0 {
				fromNodeID := task.ID
				for _, notifyName := range task.Notify {
					if handlerTask, ok := handlerMap[notifyName]; ok {
						// Use the new handler-specific ID format
						toNodeID := handlerTask.ID
						graph(sb, 1, fmt.Sprintf("%s -> %s [style=dotted, color=purple, label=\"🚨 notify\", weight=0];", fromNodeID, toNodeID))
					}
				}
			}
		}
	}
}

func renderDependency(sb *strings.Builder, dep *Dependency) {
	fromNodeID := dep.From.ID
	toNodeID := dep.To.ID

	switch dep.Type {
	case DepTypeRegister:
		graph(sb, 1, fmt.Sprintf("%s -> %s [style=dashed, color=blue, label=\"🔗 %s\"];", fromNodeID, toNodeID, dep.Label))
	case DepTypeFact:
		graph(sb, 1, fmt.Sprintf("%s -> %s [style=dashed, color=darkgreen, label=\"📊 fact\"];", fromNodeID, toNodeID))
	}
}

func connectNodes(sb *strings.Builder, fromID, toID, label string) {
	if fromID != "" && toID != "" {
		graph(sb, 2, fmt.Sprintf("%s -> %s [%s];", fromID, toID, label))
	}
}
