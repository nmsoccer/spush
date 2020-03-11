package main;
/*
*SPush is Created by nmsoccer
* more instruction could be found @https://github.com/nmsoccer/spush
*/
import (
	"fmt"
	"encoding/json"
	"os"
	"flag"
	"strings"
	"bytes"
	"net"
	"time"
	"os/exec"
	"strconv"
)

var DefaultPushTimeout int =30 //default 30s timeout
var TransProto="udp"
var ListenPort int =32808;	// dispatcher listen port
var ListenAddr=":" + strconv.Itoa(ListenPort); //dispatcher listen addr

var ConfFile string = "./conf.json";	//default conf file
var CfgDir string = "./cfg/";
var tmpl_map map[string]string = make(map[string]string);


//options
var PushAll = flag.Bool("P", false , "push all procs");
var PushSome = flag.String("p", "", "push some procs");
var CreateCfg = flag.Bool("C", false , "just create cfg");
var Verbose = flag.Bool("v", false , "verbose");
var SConfFile = flag.String("f", "", "spec conf file,default using ./conf.json");
var RemainFootPrint = flag.Bool("r", false, "remain footprint at every deployed dir");


type Proc struct {
	Name string 
	Bin []string  	
	Host string 
	HostDir string `json:"host_dir"`
	CopyCfg int `json:"copy_cfg"`
}

type ProcCfg struct {
	Name string
	CfgName string	`json:"cfg_name"`
	CfgTmpl string	`json:"cfg_tmpl"`
	TmplParam string `json:"tmpl_param"`
}


type Conf struct {
	TaskName string `json:"task"`
	DeployHost string `json:"deploy_host"`
	DeployTimeOut int `json:"deploy_timeout"`
	DeployUser string `json:"remote_user"`
	DeployPass string `json:"remote_pass"` 	
	Procs []*Proc
	ProcCfgs []*ProcCfg	`json:"proc_cfgs"`
}

type MProc struct {
	proc *Proc
	cfg_file string
}

const (
	PUSH_ING int = iota
	PUSH_SUCCESS 
	PUSH_ERR
)

type TransMsg struct {
	Mtype int `json:"msg_type"` //1:report 2:response
	Mproc string `json:"msg_proc"`
	Mresult int `json:"msg_result"` //refer const PUSH_XX
	Minfo string `json:"msg_info"`
}

type PushResult struct {
	proc *Proc
	status int	//refer const PUSH_XX
	info string
}


var conf Conf;
var proc_map = make(map[string] *MProc);

func main() {
	fmt.Println("spush starts...");
	flag.Parse();
	//check flag
	if flag.NFlag() <= 0 {
		flag.PrintDefaults();
		return;
	}
	
	if *SConfFile != "" && len(*SConfFile)>0 {
		ConfFile = *SConfFile;
	}
	
	//open conf
	file , err := os.Open(ConfFile);
	if err != nil {
		fmt.Printf("open %s failed! err:%v", ConfFile , err);
		return;
	}
	defer file.Close();
	
	//decode	
	var decoder *json.Decoder;
	decoder = json.NewDecoder(file);
	err = decoder.Decode(&conf);
	if err != nil {
		fmt.Printf("decode failed! err:%s", err);
		return;
	}
	if len(conf.Procs)<=0 {
		fmt.Printf("empty proc! nothing to do\n");
		return;
	}
	
	//check arg
	if conf.TaskName == "" || len(conf.TaskName)<=0 {
		fmt.Printf("conf.task not set ! please check\n");
		return;
	}
	
	if conf.DeployHost == "" {
		conf.DeployHost = "127.0.0.1";
	}
	
	if conf.DeployTimeOut == 0 {
		conf.DeployTimeOut = DefaultPushTimeout; //default 60s
	}
	
			
	//pp
	if !*Verbose {			
		go func() {
			for {
				fmt.Printf(".");
				time.Sleep(1e9); //1s
			}
		}()
	}		
	
	//mproc
	var mproc *MProc;
	for _ , proc := range conf.Procs {
		if proc.Host=="" || len(proc.Host)<=0 {	//default set local
			proc.Host="127.0.0.1";
		}
		mproc = new(MProc);
		mproc.proc = proc;		
		proc_map[proc.Name] = mproc;
		v_print("proc:%s bin:%v host:%s host_dir:%s copy_cfg:%d\n", proc.Name , mproc.proc.Bin , mproc.proc.Host , mproc.proc.HostDir , 
			mproc.proc.CopyCfg);		
	}
	v_print("proc_map:%v\n", proc_map);
		
	//handle option
	switch  {
	case *CreateCfg:
		fmt.Println("create cfg...");
		if len(conf.ProcCfgs)<=0 {
			fmt.Printf("nothing to do\n");
			break;
		}
		var pcfg *ProcCfg;
		for _ , pcfg = range conf.ProcCfgs {
			//proc_map[pcfg.Name].cfg_file = pcfg.CfgName;
			create_cfg(pcfg);
		}
		//fallthrough;
		break;
	case *PushAll:
		fmt.Println("push all procs");
			//1. create cfg
		var pcfg *ProcCfg;
		for _ , pcfg = range conf.ProcCfgs {
			create_cfg(pcfg);
		}
		
			//2. routine			
		ch := make(chan string)
		push_result := init_push_result(conf.Procs);		
		go check_push_result(ch , push_result);
		
		timing_ch := time.Tick(time.Second * time.Duration(conf.DeployTimeOut)); //default 30s timeout
					
			//3. gen pkg	
		var pproc *Proc;	
		for _ , pproc = range conf.Procs {
			gen_pkg(pproc);
		}
		
			//4. push result
		select {
		case <- timing_ch:
			fmt.Printf("\n----------Push <%s> Timeout----------\n" , conf.TaskName);

		case push_result := <- ch:
			fmt.Printf("\n----------Push <%s> Result---------- \n%s\n", conf.TaskName , push_result);	
		}
		print_push_result(push_result);
		break;
	case *PushSome != "":
		fmt.Printf("try to push some procs:%s\n", *PushSome);
		break;
	default:
		fmt.Println("nothing to do");
		break;
	}
	
	
	return;
}

//src like "key=value , key2=value2 , ..."
func parse_tmpl_param2(src string , result map[string]string)  int {
	if src == "" {
		return -1;
	}
	
	//split "," 
	str_list := strings.Split(src, ",");
	fmt.Printf("str_list:%v\n", str_list);
	
	//split "="
	for _ , item := range str_list {
		k_v := strings.Split(item, "=");
		v_print("key=%s value=%s\n", k_v[0] , k_v[1]);
		result[k_v[0]] = k_v[1];
	}
	fmt.Printf("result:%v\n", result);
	return 0;	
}

func parse_tmpl_param(src string , result map[string]string)  int {
	if src == "" {
		return -1;
	}
	
	//split ","
	org := []byte(src); 
	bytes_list := bytes.Split(org, []byte(","));
		
	//split "=" (item like "key=value")
	for _ , item := range bytes_list {
		k_v := bytes.Split(item, []byte("="));
		
		k_v[0] = bytes.Trim(k_v[0] , " ");
		k_v[1] = bytes.Trim(k_v[1] , " ");
		result[string(k_v[0])] = string(k_v[1]);
	}
	//fmt.Printf("result:%v\n", result);
	return 0;	
}

func gen_pkg(pproc *Proc) int {
	var _func_ = "<gen_pkg>";
		
	//pkg-dir
	//curr_time := time.Now();
	var pkg_dir = ""
	pkg_dir = fmt.Sprintf("./pkg/%s/%s/" , conf.TaskName , pproc.Name);
	//pkg_dir = fmt.Sprintf("./pkg/%s/%d-%02d-%02d/" , pproc.Name , curr_time.Year() , int(curr_time.Month()) , curr_time.Day());
	//fmt.Println(pkg_dir);
	
	//rm exist dir
	err := os.RemoveAll(pkg_dir);
	if err != nil {
		fmt.Printf("%s remove old dir:%s failed! err:%s\n", _func_ , pkg_dir , err);
		return -1;
	}
	
	//create dir
	err = os.MkdirAll(pkg_dir , 0766);
	if err != nil {
		fmt.Printf("%s create dir %s failed! err:%s\n", _func_ , pkg_dir , err);
		return -1;
	}
	
	//copy files	
	cp_arg := []string{"-f"};
	cp_arg = append(cp_arg, pproc.Bin...);
	if pproc.CopyCfg == 1 {		//copy cfg
		cp_arg = append(cp_arg , proc_map[pproc.Name].cfg_file);
	}	
	cp_arg = append(cp_arg , pkg_dir);
	v_print("exe cp %v\n" , cp_arg);
		
	cp_cmd := exec.Command("cp", cp_arg...);
	output_info := bytes.Buffer{};
	cp_cmd.Stdout = &output_info;
	err = cp_cmd.Run();
	if err != nil {
		fmt.Printf("exe cmd failed! err:%s cmd:%v\n", err , cp_cmd.Args);
		return -1;
	}
	
	//exe tool
	push_cmd := exec.Command("./tools/push.sh", conf.TaskName , pproc.Name , pproc.Host , pproc.HostDir , conf.DeployHost , strconv.Itoa(ListenPort) , 
		conf.DeployUser , conf.DeployPass);
	cmd_result := bytes.Buffer{};
	push_cmd.Stdout = &cmd_result;
	err = push_cmd.Run();
	if err != nil {
		fmt.Printf("exe cmd failed! err:%s cmd:%v\n", err , push_cmd.Args);
		return -1;
	}
	if *Verbose {
		fmt.Printf("%s\n", cmd_result.String());
	}
	return 0; 
}



func create_cfg(pcfg *ProcCfg) int {
	cfg_path := CfgDir + conf.TaskName + "/" + pcfg.Name; //cfg/$task/$proc_name/
	
	//create dir
	err := os.MkdirAll(cfg_path, 0766);
	if err != nil {
		fmt.Printf("create dir %s failed! err:%s", cfg_path , err);
		return -1;
	}
	
	//create file
	cfg_real := cfg_path + "/" + pcfg.CfgName;
	fp , err := os.OpenFile(cfg_real, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644);
	if err != nil {
		fmt.Printf("open %s failed! err:%s", cfg_real , err);
		return -1;
	}
	defer fp.Close();
	
	//read tmpl content
	if pcfg.CfgTmpl != "" && tmpl_map[pcfg.CfgTmpl] == "" {
		tmp_fp  , err := os.Open(pcfg.CfgTmpl);
		if err != nil {
			fmt.Printf("open %s failed! err:%s", pcfg.CfgTmpl , err);
			return -1;
		}
		defer tmp_fp.Close();
		
		//read
		var content []byte = make([]byte , 2048);
		n , err := tmp_fp.Read(content);
		if err != nil {
			fmt.Printf("read %s failed! err:%s", pcfg.CfgTmpl , err);
			return -1;
		}
		tmpl_map[pcfg.CfgTmpl] = string(content[:n]);
	}	
	
	//parse tmpl param
	cfg_content  := []byte(tmpl_map[pcfg.CfgTmpl]);
	if pcfg.TmplParam != "" {
		tmpl_param_map := make(map[string]string);
		res := parse_tmpl_param(pcfg.TmplParam , tmpl_param_map);
		if res != 0 {
			fmt.Printf("parse tmpl param failed! cfg:%s tmpl:%s param:%s", cfg_real , pcfg.CfgTmpl , pcfg.TmplParam);
			return -1;
		}
		
		//replace param
		for k , v := range tmpl_param_map {
			cfg_content = bytes.ReplaceAll(cfg_content, []byte("$"+k), []byte(v));
		}
						
	}
	//fmt.Printf("after parsing:%s\n len:%d\n", string(cfg_content) , len(cfg_content));
	
	//write to cfg
	_ , err = fp.Write(cfg_content[:len(cfg_content)]);
	if err != nil {
		fmt.Printf("write to %s failed! err:%s", cfg_real , err);
		return -1;
	}
	proc_map[pcfg.Name].cfg_file = cfg_real;
	fmt.Printf("create %s success!\n", cfg_real);
	return 0;
}

func init_push_result(procs []*Proc) (map[string]*PushResult) {
	push_map := make(map[string] *PushResult);
	
	for _ , pproc := range procs {
		push_map[pproc.Name] = &PushResult{status:PUSH_ING , info:"dipatching" , proc:pproc};
	}
	return push_map;
}

func print_push_result(check_map map[string]*PushResult) {
	code_converse := map[int]string {PUSH_ING:"timeout" , PUSH_SUCCESS:"success" , PUSH_ERR:"err"};
	for proc_name , presult := range check_map {
		fmt.Printf("[%s]::%s %s\n", proc_name , code_converse[presult.status] , presult.info);
	}
}

func check_push_result(c chan string , check_map map[string]*PushResult) {
	complete := 0;
	result := "ok";
	var conn *net.UDPConn;
	
	//construct check map
	v_print("map len:%d and map:%v\n", len(check_map) , check_map);
		
	//resolve addr
	my_addr , err := net.ResolveUDPAddr(TransProto, ListenAddr);
	if err != nil {
		fmt.Printf("resolve  %s failed! we may not recv push results! err:%s", ListenAddr , err);
		result = "fail";
		goto _end;
	}
	
	//listen and response
	conn , err = net.ListenUDP(TransProto, my_addr);
	if err != nil {
		fmt.Printf("listen  %s failed! we may not recv push results! err:%s", ListenAddr , err);
		result = "fail";
		goto _end;
	}
	
	//handle
	for {
		//check complete
		if complete == len(check_map) {
			result = "ok";
			break;
		}
		
		//read pkg
		recv_buff := make([]byte , 256);
		n  , peer_addr , err := conn.ReadFromUDP(recv_buff);
		if err != nil {
			fmt.Printf("recv from udp failed! err:%s\n", err);
			time.Sleep(1e9); //1s
			continue;
		}
		
		//print pkg
		recv_buff = recv_buff[:n];
		v_print("recv from %s msg:%s\n", peer_addr.String() , string(recv_buff));
		
		//decode
		var msg TransMsg;
		err = json.Unmarshal([]byte(recv_buff), &msg);
		if err != nil {
			fmt.Printf("json decode failed! err:%s\n", err);
			continue;
		} 				
		v_print("msg:%v\n", msg);
		
		//check
		if msg.Mtype != 1 {
			fmt.Printf("mst type illegal! type:%d\n", msg.Mtype);
			continue;
		}
		if check_map[msg.Mproc] == nil {
			fmt.Printf("msg proc illegal! proc:%s\n", msg.Mproc);
			continue;
		}
		
		//set status
		if check_map[msg.Mproc].status == PUSH_ING {
			check_map[msg.Mproc].status = msg.Mresult;
			check_map[msg.Mproc].info = msg.Minfo;
			complete += 1;
		}
		
		//response
		conn.WriteTo([]byte("good night"), peer_addr);		
	}
	
	
	
_end:	
	c <- result;
}

func v_print(format string , ext_arg ...interface{}) {
	if *Verbose {
		fmt.Printf(format, ext_arg...);
	}
}
