package dfsm

import (
	"fmt"
	"strings"
	"context"
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


type Definition struct {
        InitialState State        `yaml:"initial"`
        States       []State       `yaml:"states"`
        Permissions  []Permission  `yaml:"permissions"`
        Transitions  []Transition  `yaml:"transitions"`
}



type EventList struct {
	Et []EventTuple	`yaml:"events"`	
}

type EventTuple struct {
	ID string `yaml:"id"`
	Value string `yaml:"value"`
	Event string `yaml:"event"`
}


type DomainFsm struct {
	Def  Definition
	Evmap   map[string]string
	fsm  *fsm.FSM
}

func NewDomainFsm(stateyaml string,eventyaml string) (*DomainFsm,error){
	var def Definition
	var el  EventList

        statebin := []byte(stateyaml)
        err := yaml.Unmarshal(statebin, &def)
        if err != nil {
        	return nil,err
	}
        // Go言語のFSMを作成する
        fsm := fsm.NewFSM(
                def.InitialState.Name,         // 初期状態
                genEvents(def.Transitions),             // 遷移
                fsm.Callbacks{},          // コールバック（省略可能）
        )
	eventbin := []byte(eventyaml)
        err = yaml.Unmarshal(eventbin, &el)
        if err != nil {
        	return nil,err
        } 
	emap := make(map[string]string)
	for _, et := range el.Et {
		emap[et.ID+et.Value] = et.Event
	}
	df := &DomainFsm{fsm: fsm, Def: def, Evmap: emap}

        return df,nil 
}

func (df *DomainFsm) Input(id string, value string)(bool,string,error) {
	eventname,exists := df.genEvent(id,value)	
	if exists == false {
		fmt.Println("event not found.default selected")
		eventname =  "Default"
	} 
	fmt.Printf("check Permit call:[%v]\n",eventname)
	permit := df.checkPermitEvent(eventname)
	
	if permit != true {
		return permit,df.fsm.Current(),nil
	} 
	err := df.fsm.Event(context.Background(),eventname)
	if err != nil {
		return permit,df.fsm.Current(),err
	}
	return permit,df.fsm.Current(),nil
}

func (df *DomainFsm)genEvent(id string,value string) (string,bool){
	eventkey := id+value
	eventvalue := ""
	evfind := false
	if val,ok := df.Evmap[eventkey]; ok {
		eventvalue = val
		evfind = ok
	}
	return eventvalue,evfind	
}

func (df *DomainFsm)checkPermitEvent(event string)(bool){
	fmt.Printf("permission:%v\n",df.Def.Permissions)
	for _, p := range df.Def.Permissions {
		if p.Event == event {
			fmt.Println("check event found")
			for _, s := range p.Permits {
				if (s.Name == df.fsm.Current()) {
					return true
				}
			}
		}
	}	
	return false
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
