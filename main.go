package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go_fsm_marmaid/dfsm"
)

const statedata string = `
initial: 
  name: stateA
states:
- name: stateA
- name: stateB
- name: stateC
permissions:
- event: eventX
  permits:
  - name : stateA 
  - name : stateC 
- event: eventY
  permits:
  - name : stateA 
  - name : stateB 
- event: eventZ
  permits:
  - name : stateA 
  - name : stateC 
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
`

const eventdata string = `
events:
  - id : testid1
    value: testvalue1
    event: eventX
  - id : testid1
    value: testvalue2
    event: eventY
  - id : testid2
    value: testvalue1
    event: eventZ
`

type InputRes struct {
    Permit  string    `json:"permit"`
    Id      string    `json:"id"`
    Value   string    `json:"value"`
    State string `json:"state"`
} 

func main() {
	r := gin.Default()

	ds,err := dfsm.NewDomainFsm(statedata,eventdata)
	if err != nil {
		panic(err)
	}
	fmt.Println(ds.Evmap)

	var res InputRes	
		
	r.Static("/static", "./static")
	r.GET("/test/testid1", func(c *gin.Context) {
		permit,statename,err := ds.Input("testid1","testvalue1")
		if err != nil {
			panic(err)
		}
		if (permit) {
			res.Permit = "Permit"
		} else {
			res.Permit = "Deny"
		}
		res.Id = "testid1"
		res.Value = "testvalue1"
		res.State = statename	
		c.JSON(200,res)
        })
	r.GET("/mermaid", func(c *gin.Context) {
		// Mermaidコードの生成 (現在の状態をStateAとしてスタイル変更
		crcode := ds.GenCodeSrcFsm()
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


/*
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
*/
