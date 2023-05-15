package dfsm

import (
	"fmt"
	"strings"
        "github.com/looplab/fsm"
        "gopkg.in/yaml.v2"
)


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


type DomainFsm struct {
	Def  Definition
	fsm  *fsm.FSM
}

func NewDomainFsm(yamldata string) (*DomainFsm) {
	var def Definition
        yamlbin := []byte(yamldata)
        err := yaml.Unmarshal(yamlbin, &def)
        if err != nil {
                panic(err)
        }
        // Go言語のFSMを作成する
        fsm := fsm.NewFSM(
                def.InitialState.Name,         // 初期状態
                genEvents(def.Transitions),             // 遷移
                fsm.Callbacks{},          // コールバック（省略可能）
        )
	
	df := &DomainFsm{fsm: fsm, Def: def}
        return df 
}

func (df *DomainFsm) GenCodeSrcFsm()(string) {
        var sb strings.Builder
        sb.WriteString("graph TD;\n")
        for _, e := range df.Def.Transitions {
                for _, s := range e.Src {
                        sb.WriteString(fmt.Sprintf("%s --%s--> %s;\n", s.Name, e.Event, e.Dst[0].Name))
                }
        }
        sb.WriteString(fmt.Sprintf("style %s fill:#7BCCAC\n",df.fsm.Current()))
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
