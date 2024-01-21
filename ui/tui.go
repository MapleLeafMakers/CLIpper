package ui

import (
	"clipper/jsonrpcclient"
	"clipper/ui/cmdinput"
	"encoding/json"
	"github.com/bykof/gostradamus"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
)

const wordList = "ability,able,about,above,accept,according,account,across,act,action,activity,actually,add,address,administration,admit,adult,affect,after,again,against,age,agency,agent,ago,agree,agreement,ahead,air,all,allow,almost,alone,along,already,also,although,always,American,among,amount,analysis,and,animal,another,answer,any,anyone,anything,appear,apply,approach,area,argue,arm,around,arrive,art,article,artist,as,ask,assume,at,attack,attention,attorney,audience,author,authority,available,avoid,away,baby,back,bad,bag,ball,bank,bar,base,be,beat,beautiful,because,become,bed,before,begin,behavior,behind,believe,benefit,best,better,between,beyond,big,bill,billion,bit,black,blood,blue,board,body,book,born,both,box,boy,break,bring,brother,budget,build,building,business,but,buy,by,call,camera,campaign,can,cancer,candidate,capital,car,card,care,career,carry,case,catch,cause,cell,center,central,century,certain,certainly,chair,challenge,chance,change,character,charge,check,child,choice,choose,church,citizen,city,civil,claim,class,clear,clearly,close,coach,cold,collection,college,color,come,commercial,common,community,company,compare,computer,concern,condition,conference,Congress,consider,consumer,contain,continue,control,cost,could,country,couple,course,court,cover,create,crime,cultural,culture,cup,current,customer,cut,dark,data,daughter,day,dead,deal,death,debate,decade,decide,decision,deep,defense,degree,Democrat,democratic,describe,design,despite,detail,determine,develop,development,die,difference,different,difficult,dinner,direction,director,discover,discuss,discussion,disease,do,doctor,dog,door,down,draw,dream,drive,drop,drug,during,each,early,east,easy,eat,economic,economy,edge,education,effect,effort,eight,either,election,else,employee,end,energy,enjoy,enough,enter,entire,environment,environmental,especially,establish,even,evening,event,ever,every,everybody,everyone,everything,evidence,exactly,example,executive,exist,expect,experience,expert,explain,eye,face,fact,factor,fail,fall,family,far,fast,father,fear,federal,feel,feeling,few,field,fight,figure,fill,film,final,finally,financial,find,fine,finger,finish,fire,firm,first,fish,five,floor,fly,focus,follow,food,foot,for,force,foreign,forget,form,former,forward,four,free,friend,from,front,full,fund,future,game,garden,gas,general,generation,get,girl,give,glass,go,goal,good,government,great,green,ground,group,grow,growth,guess,gun,guy,hair,half,hand,hang,happen,happy,hard,have,he,head,health,hear,heart,heat,heavy,help,her,here,herself,high,him,himself,his,history,hit,hold,home,hope,hospital,hot,hotel,hour,house,how,however,huge,human,hundred,husband,idea,identify,if,image,imagine,impact,important,improve,in,include,including,increase,indeed,indicate,individual,industry,information,inside,instead,institution,interest,interesting,international,interview,into,investment,involve,issue,it,item,its,itself,job,join,just,keep,key,kid,kill,kind,kitchen,know,knowledge,land,language,large,last,late,later,laugh,law,lawyer,lay,lead,leader,learn,least,leave,left,leg,legal,less,let,letter,level,lie,life,light,like,likely,line,list,listen,little,live,local,long,look,lose,loss,lot,love,low,machine,magazine,main,maintain,major,majority,make,man,manage,management,manager,many,market,marriage,material,matter,may,maybe,me,mean,measure,media,medical,meet,meeting,member,memory,mention,message,method,middle,might,military,million,mind,minute,miss,mission,model,modern,moment,money,month,more,morning,most,mother,mouth,move,movement,movie,Mr,Mrs,much,music,must,my,myself,n't,name,nation,national,natural,nature,near,nearly,necessary,need,network,never,new,news,newspaper,next,nice,night,no,none,nor,north,not,note,nothing,notice,now,number,occur,of,off,offer,office,officer,official,often,oh,oil,ok,old,on,once,one,only,onto,open,operation,opportunity,option,or,order,organization,other,others,our,out,outside,over,own,owner,page,pain,painting,paper,parent,part,participant,particular,particularly,partner,party,pass,past,patient,pattern,pay,peace,people,per,perform,performance,perhaps,period,person,personal,phone,physical,pick,picture,piece,place,plan,plant,play,player,PM,point,police,policy,political,politics,poor,popular,population,position,positive,possible,power,practice,prepare,present,president,pressure,pretty,prevent,price,private,probably,problem,process,produce,product,production,professional,professor,program,project,property,protect,prove,provide,public,pull,purpose,push,put,quality,question,quickly,quite,race,radio,raise,range,rate,rather,reach,read,ready,real,reality,realize,really,reason,receive,recent,recently,recognize,record,red,reduce,reflect,region,relate,relationship,religious,remain,remember,remove,report,represent,Republican,require,research,resource,respond,response,responsibility,rest,result,return,reveal,rich,right,rise,risk,road,rock,role,room,rule,run,safe,same,save,say,scene,school,science,scientist,score,sea,season,seat,second,section,security,see,seek,seem,sell,send,senior,sense,series,serious,serve,service,set,seven,several,sex,sexual,shake,share,she,shoot,short,shot,should,shoulder,show,side,sign,significant,similar,simple,simply,since,sing,single,sister,sit,site,situation,six,size,skill,skin,small,smile,so,social,society,soldier,some,somebody,someone,something,sometimes,son,song,soon,sort,sound,source,south,southern,space,speak,special,specific,speech,spend,sport,spring,staff,stage,stand,standard,star,start,state,statement,station,stay,step,still,stock,stop,store,story,strategy,street,strong,structure,student,study,stuff,style,subject,success,successful,such,suddenly,suffer,suggest,summer,support,sure,surface,system,table,take,talk,task,tax,teach,teacher,team,technology,television,tell,ten,tend,term,test,than,thank,that,the,their,them,themselves,then,theory,there,these,they,thing,think,third,this,those,though,thought,thousand,threat,three,through,throughout,throw,thus,time,to,today,together,tonight,too,top,total,tough,toward,town,trade,traditional,training,travel,treat,treatment,tree,trial,trip,trouble,true,truth,try,turn,TV,two,type,under,understand,unit,until,up,upon,us,use,usually,value,various,very,victim,view,violence,visit,voice,vote,wait,walk,wall,want,war,watch,water,way,we,weapon,wear,week,weight,well,west,western,what,whatever,when,where,whether,which,while,white,who,whole,whom,whose,why,wide,wife,will,win,wind,window,wish,with,within,without,woman,wonder,word,work,worker,world,worry,would,write,writer,wrong,yard,yeah,year,yes,yet,you,young,your,yourself"

type LogEntry struct {
	Timestamp gostradamus.DateTime
	Message   string
}

type LogContent struct {
	tview.TableContentReadOnly
	entries []LogEntry
	table   *tview.Table
}

func (l *LogContent) Write(message string) {
	lines := cmdinput.WordWrap(cmdinput.Escape(message), 999) //arbitrary number to force it to only split on actual linebreaks for now
	for _, line := range lines {
		l.entries = append(l.entries, LogEntry{gostradamus.Now(), line})
	}
	l.table.ScrollToEnd()
}

func (l LogContent) GetCell(row, column int) *tview.TableCell {

	switch column {
	case 0:
		return tview.NewTableCell(l.entries[row].Timestamp.Format(" hh:mma")).SetBackgroundColor(tcell.NewRGBColor(64, 64, 64))
	case 1:
		return tview.NewTableCell(" " + l.entries[row].Message)
	}
	return nil
}

func (l LogContent) GetRowCount() int {
	return len(l.entries)
}

func (l LogContent) GetColumnCount() int {
	return 2
}

type TUI struct {
	App          *tview.Application
	Root         *tview.Grid
	Input        *cmdinput.InputField
	Output       *LogContent
	RpcClient    *jsonrpcclient.Client
	TabCompleter cmdinput.TabCompleter
	State        map[string]interface{}
	Settings     Settings
	HostHeader   *tview.TextView
}

func NewTUI(rpcClient *jsonrpcclient.Client) *TUI {
	tui := &TUI{
		RpcClient: rpcClient,
	}

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tui.Settings = AppSettings
	tui.buildInput()
	tui.buildOutput(100)
	tui.buildWindow()
	tui.buildLeftPanel()
	tui.App = tview.NewApplication().SetRoot(tui.Root, true).EnableMouse(true)
	go tui.loadPrinterInfo()
	go tui.loadGcodeHelp()
	go tui.subscribe()
	go tui.handleIncoming()
	return tui
}

func (tui *TUI) buildInput() {
	tui.Input = cmdinput.NewInputField().SetPlaceholder("Enter GCODE Commands or / commands").SetLabel("> ").SetLabelStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true))

	tui.TabCompleter = cmdinput.NewTabCompleter()
	tui.TabCompleter.RegisterCommand("/set", Command_Set{})
	tui.TabCompleter.RegisterCommand("/settings", Command_Settings{})
	tui.TabCompleter.RegisterCommand("/quit", Command_Quit{})
	tui.TabCompleter.RegisterCommand("/rpc", Command_RPC{})

	tui.Input.SetAutocompleteFunc(func(currentText string) (entries []string) {
		ctx := cmdinput.CommandContext{
			"tui": tui,
			"raw": currentText,
		}
		return tui.TabCompleter.AutoComplete(currentText, tui.Input.GetCursor(), ctx)
	})

	tui.Input.SetAutocompletedFunc(func(text string, index, source int) bool {
		closeMenu, fullText, cursorPos := tui.TabCompleter.OnAutoCompleted(text, index, source)
		tui.Input.SetText(fullText)
		tui.Input.SetCursor(cursorPos)
		return closeMenu
	})

	tui.Input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyPgDn {
			tui.Output.table.InputHandler()(event, func(p tview.Primitive) {})
		} else {
			return event
		}
		return nil
	})

	tui.Input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			ctx := cmdinput.CommandContext{
				"tui": tui,
				"raw": tui.Input.GetText(),
			}
			err := tui.TabCompleter.Parse(tui.Input.GetText(), ctx)
			if err == nil {
				log.Println("Executing", ctx)
				cmd, ok := ctx["cmd"]
				// ew
				if ok {
					cmd2, ok2 := cmd.(cmdinput.Command)
					if ok2 {
						go cmd2.Call(ctx)
					}
				} else {
					//not a registered command, send it as gcode.
					go (func() { NewGcodeCommand("", "").Call(ctx) })()
				}
				tui.Input.Clear()
			} else if err.Error() == "NoInput" {

			} else {
				panic(err)
			}
		default:
		}
	})
}

func (tui *TUI) buildLeftPanel() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	tui.HostHeader = tview.NewTextView().SetTextAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow)
	tui.HostHeader.SetBackgroundColor(tcell.ColorDarkCyan)
	tui.Root.AddItem(flex, 0, 0, 1, 1, 0, 0, false)
	flex.AddItem(tui.HostHeader, 1, 1, false)
	flex.AddItem(tview.NewBox().SetBorder(true), 0, 2, false)
}

func (tui *TUI) buildOutput(numLines int) {

	output := tview.NewTable()
	ts := gostradamus.Now()
	lines := make([]LogEntry, numLines)
	i := 0
	for i = 0; i < numLines-6; i++ {
		lines[i] = LogEntry{ts, ""}
	}
	lines[i+0] = LogEntry{ts, "[yellow]   ________    ____                     "}
	lines[i+1] = LogEntry{ts, "[yellow]  / ____/ /   /  _/___  ____  ___  _____"}
	lines[i+2] = LogEntry{ts, "[yellow] / /   / /    / // __ \\/ __ \\/ _ \\/ ___/"}
	lines[i+3] = LogEntry{ts, "[yellow]/ /___/ /____/ // /_/ / /_/ /  __/ /    "}
	lines[i+4] = LogEntry{ts, "[yellow]\\____/_____/___/ .___/ .___/\\___/_/     "}
	lines[i+5] = LogEntry{ts, "[yellow]              /_/   /_/                 "}

	tui.Output = &LogContent{
		table:   output,
		entries: lines,
	}
	output.SetContent(tui.Output)
	output.ScrollToEnd()
}

func (tui *TUI) buildWindow() {
	tui.Root = tview.NewGrid().
		SetRows(0, 1).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(tui.Input, 1, 0, 1, 2, 0, 0, true).
		AddItem(tui.Output.table, 0, 1, 1, 1, 0, 0, false)
}

func (tui *TUI) subscribe() {
	resp, err := tui.RpcClient.Call("printer.objects.subscribe", map[string]interface{}{
		"objects": map[string]interface{}{
			"print_stats":  nil,
			"idle_timeout": nil,
			"gcode_move":   nil,
		},
	})
	if err != nil {
		panic(err)
	}
	state, _ := resp.(map[string]interface{})
	tui.App.QueueUpdateDraw(func() {
		tui.State = state
	})
}

func (tui *TUI) handleIncoming() {
	for {
		logIncoming, _ := AppSettings["logIncoming"].(bool)
		incoming := <-tui.RpcClient.Incoming
		switch incoming.Method {
		case "notify_status_update":
			status := incoming.Params[0].(map[string]interface{})
			if logIncoming {
				out, _ := json.MarshalIndent(status, "", " ")
				tui.App.QueueUpdateDraw(func() {
					tui.Output.Write(string(out))
				})
			}
		case "notify_gcode_response":
			tui.App.QueueUpdateDraw(func() {
				for _, line := range incoming.Params {
					tui.Output.Write(line.(string))
				}
			})
		default:
			if logIncoming && incoming.Method != "notify_proc_stat_update" {
				out, _ := json.MarshalIndent(incoming, "", " ")
				tui.App.QueueUpdateDraw(func() {
					tui.Output.Write(string(out))
				})
			}
		}
	}
}

func (tui *TUI) loadPrinterInfo() {
	resp, err := tui.RpcClient.Call("printer.info", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	info, _ := resp.(map[string]interface{})
	tui.App.QueueUpdateDraw(func() {
		tui.HostHeader.SetText(info["hostname"].(string))
	})
}

func (tui *TUI) loadGcodeHelp() {
	resp, err := tui.RpcClient.Call("printer.gcode.help", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	tui.App.QueueUpdate(func() {
		for k, help := range resp.(map[string]interface{}) {
			tui.TabCompleter.RegisterCommand(k, NewGcodeCommand(k, help.(string)))
		}
	})
}

func dumpToJson(obj any) string {
	out, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return "<error>"
	}
	return string(out)
}
