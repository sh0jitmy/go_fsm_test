package dfsm

import (
	"fmt"
	"strings"
	"context"
	"errors"
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
}

type AdhocFsm struct {
	fsm  *fsm.FSM
	evmap   map[string]string
	permissions []Permission
}

func NewDomainFsm(stateyaml string,eventyaml string) (*DomainFsm,error){
	var def Definition
	var el  EventList

        statebin := []byte(stateyaml)
        err := yaml.Unmarshal(statebin, &def)
        if err != nil {
        	return nil,err
	}
	eventbin := []byte(eventyaml)
        err = yaml.Unmarshal(eventbin, &el)
        if err != nil {
        	return nil,err
        } 
	emap := make(map[string]string)
	for _, et := range el.Et {
		emap[et.ID+et.Value] = et.Event
	}
	df := &DomainFsm{Def: def, Evmap: emap}

        return df,nil 
}



func (df *DomainFsm) NewAdhocFsm(instate string) (*AdhocFsm,error){
	var initstate = ""

	if instate == "" {
		initstate = df.Def.InitialState.Name         // 初期状態
	}else {
		initstate = instate
	}	

	if df.checkState(initstate) != true {
		return nil,errors.New("instate is invalid")
	} 

	// Go言語のFSMを作成する
	fsm := fsm.NewFSM(
		initstate,
		genEvents(df.Def.Transitions),             // 遷移
		fsm.Callbacks{},          // コールバック（省略可能）
        )
	af := &AdhocFsm{fsm: fsm, evmap:df.Evmap, permissions:df.Def.Permissions}
	return af,nil
}

func (df *DomainFsm) checkState(state string)(bool) {
        for _, s := range df.Def.States {
		if s.Name == state {
			return true
		}
	}
	return false
}

func (df *DomainFsm) GenCodeSrcFsm(af *AdhocFsm)(string) {
        var sb strings.Builder
        sb.WriteString("graph TD;\n")
        for _, e := range df.Def.Transitions {
                for _, s := range e.Src {
                        sb.WriteString(fmt.Sprintf("%s --%s--> %s;\n", s.Name, e.Event, e.Dst[0].Name))
                }
        }
        sb.WriteString(fmt.Sprintf("style %s fill:#7BCCAC\n",af.fsm.Current()))
        str := sb.String() 
        fmt.Println(str)
        return str
}

func (af *AdhocFsm) Input(id string, value string)(bool,string,error) {
	eventname,exists := af.decodeEvent(id,value)	
	if exists == false {
		fmt.Println("event not found.default selected")
		eventname =  "Default"
	} 
	fmt.Printf("check Permit call:[%v]\n",eventname)
	permit := af.checkPermitEvent(eventname)
	
	if permit != true {
		return permit,af.fsm.Current(),nil
	} 
	err := af.fsm.Event(context.Background(),eventname)
	if err != nil {
		return permit,af.fsm.Current(),err
	}
	return permit,af.fsm.Current(),nil
}

func (af *AdhocFsm)decodeEvent(id string,value string) (string,bool){
	eventkey := id+value
	eventvalue := ""
	evfind := false
	if val,ok := af.evmap[eventkey]; ok {
		eventvalue = val
		evfind = ok
		return eventvalue,evfind	
	} 
	//if notfound,value  wild card routine
	eventkey = id+"*"
	if val,ok := af.evmap[eventkey]; ok {
		eventvalue = val
		evfind = ok
	} 
	return eventvalue,evfind	
}

func (af *AdhocFsm)checkPermitEvent(event string)(bool){
	for _, p := range af.permissions {
		if p.Event == event {
			fmt.Println("check event found")
			for _, s := range p.Permits {
				if (s.Name == af.fsm.Current()) {
					return true
				}
			}
		}
	}	
	return false
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
