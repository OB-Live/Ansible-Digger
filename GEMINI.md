Below is the **final consolidated specification document** you requested.

This is written as a production-grade implementation contract for a code-generating AI.

---

# GEMINI.md

# Project: **Ansible Digger**

Binary: `adigger`

---

# 1. Objective

**Ansible Digger** parses an Ansible `playbook.yaml`, enriches it with installed role metadata, and generates a **Graphviz DOT** digraph designed for:

* Conversion to SVG
* Embedding into an HTML page
* Interactive exploration via **D3.js**
* Zooming into roles and modules
* Clickable nodes and clusters

The graph must be:

* Rich (high semantic density)
* Structurally layered
* Top-to-bottom
* Visually expressive
* Not excessively deep (no recursive explosion of internals)

The DOT output must be optimized for:

* SVG interactivity
* DOM manipulation
* CSS styling
* D3 cluster collapsing/expanding

---

# 2. Technical Stack

* Go ≥ 1.21
* YAML: `gopkg.in/yaml.v3`
* No Graphviz Go bindings
* Generate DOT only
* SVG conversion done externally
* Designed for D3.js post-processing

Project layout:

```
adigger/
├── go.mod
├── main.go
└── adigger/
    ├── parser.go
    ├── model.go
    ├── renderer.go
    └── ...
```

---

# 3. Rendering Principles

## Global Graph

```dot
digraph "Ansible Digger" {
    rankdir=TB;
    compound=true;
    fontname="Helvetica";
    node [fontname="Helvetica"];
    edge [fontname="Helvetica"];
}
```

Strict top-to-bottom execution.

Clusters must be:

* Collapsible (cluster IDs stable & deterministic)
* Clickable (URL attribute set)
* Tooltip-enabled
* CSS-styleable

Each node must include:

```
id=
class=
tooltip=
URL=
```

For D3 targeting.

---

# 4. Structural Layout

## Play Hierarchy

```
cluster_play_N
    cluster_pre_tasks
        cluster_task_X
    cluster_roles
        cluster_role_X
    cluster_tasks
        cluster_task_X
    cluster_handlers
```

---

# 5. Node Semantics

## Emoji System

| Concept        | Emoji |
| -------------- | ----- |
| Play           | 🎭    |
| Host           | 🖥    |
| PreTasks       | 🔧    |
| Tasks          | ⚙     |
| Roles          | 📦    |
| Handler        | 🚨    |
| Vars           | 🧾    |
| Register       | 🔗    |
| Become         | 🔒    |
| Dependency     | ⬅     |
| Block          | 📚    |
| Rescue         | 🛟    |
| Always         | ♾     |
| Fact           | 📊    |
| Parallel       | 🧵    |
| Critical       | 🔥    |
| Builtin module | 🅰    |

---

# 6. Task Subgraph Model

Each task is:

```
subgraph cluster_task_P_T {
    main_node
    vars_input_node
    previous_input_node
    fact_input_nodes
}
```

Main node includes:

* Name
* Module
* Args
* When
* Tags
* 🔒 if become
* Tooltip = raw YAML
* URL = "#task-playX-taskY"

---

# 7. Conditional Execution Branching

If `when:` exists:

Create diamond condition node:

```
shape=diamond
label="Condition"
```

Edges:

```
condition -> task [label="true"]
condition -> skip_dummy [style=dashed label="false"]
```

False branch connects to next sequential node.

This visually models branch.

---

# 8. Handler Notification Flow

If task has:

```
notify:
  - restart nginx
```

Create edge:

```
task -> handler_node
    [color=purple
     style=dotted
     label="🚨 notify"]
```

Handlers placed inside `cluster_handlers`.

---

# 9. Block / Rescue / Always Modeling

For:

```
block:
rescue:
always:
```

Create cluster:

```
cluster_block_X
    cluster_block_tasks
    cluster_rescue_tasks
    cluster_always_tasks
```

Edges:

* Block tasks sequential
* Failure edge → rescue cluster
* Completion edge → always cluster

Failure edge:

```
color=red
style=dashed
label="failure"
```

---

# 10. Parallel Execution Strategy

If play has:

```
strategy: free
```

or

```
serial:
```

Create parallelism indicator node:

```
🧵 Parallel Strategy: free
```

Parallel branches visually rendered:

Multiple edges from play node to first tasks using:

```
style=dashed
constraint=false
```

Serial groups visually separated with subtle cluster grouping.

---

# 11. Critical Path Coloring

Compute longest execution chain ignoring conditional skips.

Algorithm:

* DAG longest path
* Exclude dashed conditional edges
* Exclude notify edges

Critical path edges styled:

```
color=red
penwidth=2
label="🔥 critical"
```

Critical nodes background slightly tinted.

---

# 12. Fact Dependency Graphing

If tasks reference:

```
ansible_facts
hostvars
group_names
```

Create fact node:

```
📊 fact_name
shape=ellipse
fontsize=10
```

Edge:

```
fact_node -> task
color=darkgreen
style=dashed
label="fact"
```

Facts shared across tasks reuse same node.

---

# 13. Register Data Flow

If:

```
register: myvar
```

Producer:

Create invisible output marker.

Consumer detection:

Search in:

* args
* when
* vars
* loop
* until

Edge:

```
producer -> consumer
    [color=blue
     style=dashed
     label="🔗 myvar"]
```

Multiple allowed.

---

# 14. Vars as Input Nodes

Tiny ellipse:

```
🧾 key=value
```

Edge into main task node.

---

# 15. Role Expansion

Locate roles in:

* ./roles/
* ~/.ansible/roles/
* /etc/ansible/roles/

Parse:

```
tasks/main.yml
vars/main.yml
handlers/main.yml
templates/
files/
meta/
defaults/
```

Role subgraph:

```
cluster_role_nginx {
    main_role_node
    tasks_input_node
    vars_input_node
    handlers_input_node
    templates_input_node
    files_input_node
}
```

Only list names as input nodes.

No deep expansion.

---

# 16. Factoring for D3.js

Each cluster must include:

```
id="cluster_play_0"
class="play"
URL="#play-0"
```

Nodes:

```
id="task_0_3"
class="task module-apt"
```

Modules must add CSS class:

```
class="task module-copy"
```

So D3 can zoom by module type.

---

# 17. Expanded Go Model

Task must include:

* Loop
* WithItems
* WithDict
* DelegateTo
* IgnoreErrors
* ChangedWhen
* FailedWhen
* CheckMode
* Environment
* Notify
* Retries
* Delay
* Until
* RunOnce
* Block
* Rescue
* Always
* Async
* Poll
* Throttle
* Diff
* NoLog

Play must include:

* Vars
* VarsFiles
* Handlers
* GatherFacts
* Strategy
* Serial
* MaxFailPercentage
* ForceHandlers
* AnyErrorsFatal
* Order
* Connection

Role must include:

* Tags
* When
* Defaults
* Meta
* AllowDuplicates
* Public

---

# 18. Unused YAML Logging

Parser must:

* Track encountered keys
* Maintain whitelist
* Log unused keys:

```
[WARN] Unused task key: throttle (play 0 task 3)
```

Do not fail.

---

# 19. CLI

```
adigger -input playbook.yaml -output graph.dot -roles-path ./roles -verbose
```

---

# 20. Graph Philosophy

Graph must be:

* Rich in metadata
* Moderate depth
* Collapsible
* Clickable
* Zoomable
* Visually layered
* Execution + data flow hybrid
* Suitable for D3-driven exploration

---

# 21. Deliverables Required

Generate:

1. Full Go project
2. Example playbook
3. Generated DOT example
4. Sample HTML with D3 zoom & collapse
5. Architecture explanation
6. Build instructions

No theoretical essay.
Provide compilable, structured code.

---

If desired next iteration can introduce:

* Inventory host grouping visualization
* Network topology overlay
* Time estimation modeling
* Drift detection overlay
* Security risk heatmap
* RBAC exposure analysis

Specify direction if expanding further.
