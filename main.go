package main

import (
	"fmt"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/looplab/fsm"
	"gopkg.in/yaml.v2"
)

const statedata string = `
initial: 
  name: stateA
states:
- name: stateA
- name: stateB
- name: stateC
Permission:
- event: eventX
  permits:
  - name :stateA 
  - name :stateC 
- event: eventY
  permits:
  - name :stateA 
  - name :stateB 
- event: eventZ
  permits:
  - name :stateA 
  - name :stateC 
transitions:
- event: eventX
  dst: 
  - name: stateB
  src: 
  - name : stateA 
  - name : stateB 
- event: eventY
  dst: 
  - name: stateC
  src: 
  - name: stateB
  - name : stateC 
- event: eventZ
  dst: 
  - name: stateA
  src: 
  - name: stateC
callbacks:
- event:  eventX
  cbtype: before 
  fname : eventXcall
`

type State struct {
	Name string `yaml:"name"`
}


type Permission struct {
	Event  string `yaml:"event"`
	Permits []State `yaml:"permits"`
}

type Transition struct {
	Event string `yaml:"event"`
	Dst   []State `yaml:"dst"`
	Src   []State `yaml:"src"`
}

type Callback struct {
	Event string `yaml:"event"`
	CType string `yaml:"ctype"`
	Fname string `yaml:"fname"`
}

type Definition struct {
	InitialState State        `yaml:"initial"`
	States       []State       `yaml:"states"`
	Permissions  []Permission  `yaml:"permissions"`
	Transitions  []Transition  `yaml:"transitions"`
	Callbacks    []Callback   `yaml:"callbacks"`
}


func main() {
	r := gin.Default()

	gfsm ,def:= makeFsm(statedata)
	
	r.Static("/static", "./static")
	r.GET("/mermaid", func(c *gin.Context) {
		// Mermaidコードの生成 (現在の状態をStateAとしてスタイル変更
		crcode := genCodeSrcFsm(gfsm,def.Transitions)
		html := `
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="utf-8">
				<title>Mermaid Example</title>
				<script src="http://localhost:8580/static/js/mermaid.min.js"></script>
				<script>mermaid.initialize({startOnLoad:true});</script>
			</head>
			<body>
				<div class="mermaid">
				%s
				</div>
			</body>
			</html>
		`

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, html, crcode)
	})

	r.Run(":8580")
}



func makeFsm(yamlstr string)(*fsm.FSM,Definition) {
	var def Definition
	yamlbin := []byte(yamlstr)
	err := yaml.Unmarshal(yamlbin, &def)
	if err != nil {
		panic(err)
	}
	// Go言語のFSMを作成する
	genfsm := fsm.NewFSM(
		def.InitialState.Name,         // 初期状態
		genEvents(def.Transitions),             // 遷移
		fsm.Callbacks{},          // コールバック（省略可能）
	)
	return genfsm,def
}	

func genCodeSrcFsm(f *fsm.FSM,t []Transition)(string) {
	var sb strings.Builder
	sb.WriteString("graph TD;\n")
	for _, e := range t {
		for _, s := range e.Src {
			sb.WriteString(fmt.Sprintf("%s --%s--> %s;\n", s.Name, e.Event, e.Dst[0].Name))
		}
	}
	sb.WriteString(fmt.Sprintf("style %s fill:#7BCCAC\n",f.Current()))
	str := sb.String() 
	fmt.Println(str)
	return str 
} 


func getSrc(slist []State)([]string) {
	var namelist []string
	for _, s := range slist {
		namelist = append(namelist,s.Name)
	}
	return namelist
}

func genEvents(tslist []Transition)([]fsm.EventDesc) {
	var desc []fsm.EventDesc
	var tmp fsm.EventDesc
	for _, ts := range tslist {
		tmp.Name = ts.Event		
		tmp.Src = getSrc(ts.Src)		
		tmp.Dst = ts.Dst[0].Name
		desc = append(desc,tmp)	
	}
	return desc
}

