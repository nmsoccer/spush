# SPush
A Simple Push tool,一款简单的分发工具.主要目的是将不同的文件或目录分发到不同的机器。 

### 主要特点
* **多重分发**： 支持一次分发多个任务到多台机器及目录  
* **配置简单**： 使用一个简单的json格式的配置文件进行即可使用
* **较少依赖**： 尽量减少对环境的特殊依赖，除了运行分发机器需要部署go环境以外，目标机器只需要普通的unix/linux开发环境即可
* **结果反馈**： 为了保证分发的正确性，所有的分发任务均有分发结果反馈告知是否成功或发生错误
* **配置生成**： 在较大型的比如游戏项目中需要为不同进程编写对应的配置文件，这里支持根据模板通过进程名自动生成相关的配置文件
* **指令执行**： 为了方便一些部署工作，支持针对每个特定的分发过程在分发结束后执行简单的指令  
  _想到了再写_  
  
### 安装
 * 依赖
   * 在分发机器上需要安装golang  
   
 * 安装
   * 下载spush到本地并解压
   * 进入spush主目录./init.sh初始化
   * go build spush.go
   * ./spush xxxx 即可使用

### 命令选项
  * `-C`: 仅仅为分发任务生成各分发任务自有的配置文件(如果有设置)
  * `-P`: 推送并执行配置里所有定义的分发任务。这里会先执行-C选项的内容
  * `-p xxx`: 根据任务名选择执行配置里某些分发任务,符合一般的go正则规范即可
  * `-f conf_path`: 指定配置文件位置  
  * `-r`: 部署成功之后在部署机器的目标目录留下部署的痕迹，包括使用的工具及相关文件,默认删除
  * `-v`: 详细打印执行过程 默认关闭
  
### 基本配置
简单的配置文件参见demo/simple_copy/simple_copy.json，如下所示：   
```
{
 "task":"simple_copy" , 
 "deploy_host":"",
 "deploy_timeout":10, 
 "remote_user":"" ,
 "remote_pass":"" ,
 "procs":[
   {"name":"cpy1" , "bin":["/etc/profile.d/" , "./init.sh"] , "host":"" , "host_dir":"/home/nmsoccer/spush_demo/simple_copy/cpy1" , "copy_cfg":0 , "cmd":""} , 
   {"name":"cpy2" , "bin":["/etc/passwd" , "./count.sh"] , "host":"127.0.0.1" , "host_dir":"/home/nmsoccer/spush_demo/simple_copy/cpy2" , "copy_cfg":0 , "cmd":"./count.sh cs suomei"},
   {"name":"cpy3" , "bin":["/etc/passwd"] , "host":"" , "host_dir":"/home/nmsoccer/spush_demo/simple_copy/misc/" }, 
   {"name":"cpy4" , "bin":["./count.sh"] , "host":"" , "host_dir":"/home/nmsoccer/spush_demo/simple_copy/misc/"}
  ] ,
  
  
  "proc_cfgs":[
  ]
}
```
这是一个最简单的拷贝配置，但包含了配置文件的主要内容，除了"proc_cfg"，该选项可以为各个分发生成配置文件，留待后续说明

* 【task】: **必须**：该选项为每个配置文件指定一个task名，用于在不同的配置文件进行区分
* 【deploy_host】: **可选**：指定该分发机器的IP地址，如果全部为本地分发则可以不填或填为127.0.0.1。如果分发到不同机器则需要填写外网地址或者内部网络所在地址，否则无法收到上报信息
* 【deploy_timeout】: **可选**：指定分发超时时间，单位为秒。如果不设置则使用默认时间30秒
* 【remote_user】: **异机部署必须**：指定用于同步到部署机器的用户
* 【remote_pass】: **异机部署必须，除非配置了信任关系**：该工具使用expect脚本来进行文件同步，所以需要remote_user下使用remote_pass密码用于文件拷贝和指令执行
* 【procs】：**必须**：指定复制不同文件到不同地点的不同分发任务：  
  * 【name】: **必须**：分发任务的名字，在每个配置文件中必须唯一
  * 【bin】: **必须**：复制的源文件及目录，多个需要用','分割。注意不能使用\*等正则符，路径里的~符号也不能解析为家目录
  * 【host】: **可选**：目标主机的IP地址，如果本机部署则可以不填或者为127.0.0.1 
  * 【host_dir】:**必须**：部署到目标机器上的目录，需要填写绝对路径
  * 【copy_cfg】: **可选**：是否需要拷贝由下面proc_cfg里生成的配置文件，0：不拷贝;1:拷贝。如果proc_cfg里没有填写也不会拷贝到 
  * 【cmd】: **可选**：在该分发任务部署成功之后可以选择执行的命令。注意该命令的执行目录是在host_dir目录。比如设置./xx.sh则执行的是host_dir/xx.sh
  
* 【proc_cfgs】: **可选**：对填写的每个分发任务生成对应配置文件，普通的拷贝工作可以不使用该选项

### 基本演示 
这里使用simple_copy来进行

* 进入demo/simple_copy/目录 执行./init.sh进行初始化工作
* 修改./simple_copy.json配置文件里的host_dir为本机的有效目录
* 生成各任务配置：
  ```
  ./spush -C -f ./simple_copy.json 
  spush starts...
  create cfg...
  nothing to do
  ```
  因为没有proc_cfgs进行配置文件的设置，所以这里仿佛什么都没有发生

* 推送：
  ```
  /spush -P  -f demo/simple_copy.json 
  spush starts...
  push all procs

  :0/0
  .
  ----------Push <simple_copy> Result---------- 
  ok
  .
  [4/4]
  [cpy2]::success 
  [cpy3]::success 
  [cpy4]::success 
  [cpy1]::success
  ```
  推送成功，这里可以显示4项任务都OK鸟。我们可以去看一看目标的目录是否欧杰把客
  ```
  tree /home/nmsoccer/spush_demo/
  /home/nmsoccer/spush_demo/
  spush_demo/
  `-- simple_copy
      |-- cpy1
      |   |-- init.sh
      |   `-- profile.d
      |       |-- 256term.csh
      |       |-- 256term.sh
      |       |-- abrt-console-notification.sh
      |       |-- bash_completion.sh
      |       |-- colorgrep.csh
      |       |-- colorgrep.sh
      |       |-- colorls.csh
      |       |-- colorls.sh
      |       |-- csh.local
      |       |-- lang.csh
      |       |-- lang.sh
      |       |-- less.csh
      |       |-- less.sh
      |       |-- sh.local
      |       |-- vim.csh
      |       |-- vim.sh
      |       |-- which2.csh
      |       `-- which2.sh
      |-- cpy2
      |   |-- count.info
      |   |-- count.sh
      |   `-- passwd
      `-- misc
          |-- count.sh
          `-- passwd

    ```
  看了下都是OK的了。其中cpy3和cpy4将文件部署到了相同的目录misc。这里已经把痕迹都删除了，如果想要查看部署后的遗留可以加上-r选项。其中cpy2的分发在分发成功之后执行了./count.sh cs suomei，我们可以检查一下count.sh脚本
  ```
  cat count.sh
  #!/bin/bash
  log="./count.info"
  ts=`date +"%F %T"`
  echo $ts >> $log
  echo "$1 add $2" >> $log 
  count=1
  while [[ $count -le 20 ]]
  do
    echo "it is :${count}"  >> $log  
    let count=count+1
    sleep 1
  done
  ```
  会在当前目录打印20个数字写入count.info。我们可以看到count.info的存在，并且执行成功：
  ```
  2020-03-12 20:21:35
  cs add suomei
  it is :1
  it is :2
  it is :3
  it is :4
  ...
  ```
  
  
 * 查看日志 可以在/tmp/spush/$task/$proc/log里检查到在部署机器上执行的情况 如下：
   ```
   tree /tmp/spush/
   /tmp/spush/
   `-- simple_copy
       |-- cpy1
       |   `-- log
       |-- cpy2
       |   `-- log
       |-- cpy3
       |   `-- log
       `-- cpy4
           `-- log
   ```
   都是这样的层次。查看一下内容
  ```
  cat /tmp/spush/simple_copy/cpy1/log
  -----------------------------
   >>running on 2020-03-12 20:21:35
  ./peer_exe.sh l cpy1 127.0.0.1 32808 y [:] simple_copy 
  check md5 success
  try to run ./report
  try to send msg to 127.0.0.1:32808 proc:cpy1 stat:1 info:
  msg is:{"msg_type":1,"msg_proc":"cpy1","msg_result":1,"msg_info":""}
  [good night]
  report finish!
  deploy finish
  ```
  
  * 选择推送 只选择推送cpy2,cpy3,cpy4:
  ```
  ./spush -f demo/simple_copy.json -p "cpy[2-4]"

  ++++++++++++++++++++spush (2020-03-20 21:05:39)++++++++++++++++++++
  push some procs:cpy[2-4]
  matched procs num:3
  create cfg:0/0
  ----------Push <simple_copy> Result---------- 
  ok
  [3/3]
  [cpy2]::success 
  [cpy3]::success 
  [cpy4]::success 
  
  ```
  
### 进阶配置
  在游戏项目中，经常会部署多种类多实例进程。每种不同的进程会拥有不同的配置文件；同时每种进程可能拉起多个实例，每个实例拥有大部分相同的配置，只有部分参数彼此不同。为了解决这种情况，该工具也支持为每一个在文件里设置了的进程生成各自对应的配置文件。配置文件由模板+参数共同构成，模板是同类进程共有的静态数据；参数用于填充模板内的占位符。参考下面一个例子：
  ```
  {
  "task":"simple_game" , 
  "deploy_host":"10.161.37.100" ,
  "deploy_timeout":60, 
  "remote_user":"nmsoccer" ,
  "remote_pass":"****" ,
  
    "procs":[
    {"name":"conn_serv-1" , "bin":["./bin/conn_serv/conn_serv"] , "host":"127.0.0.1" , "host_dir":"/home/nmsoccer/sg/conn_serv" , "copy_cfg":1 , "cmd":"./conn_serv -D"},
    {"name":"conn_serv-2" , "bin":["./bin/conn_serv/conn_serv"] , "host":"10.161.37.104" , "host_dir":"/home/nmsoccer/sg/conn_serv" , "copy_cfg":1 , "cmd":"./conn_serv -D"},
	{"name":"logic_serv-1" , "bin":["./bin/logic_serv/logic_serv"] , "host":"127.0.0.1" , "host_dir":"/home/nmsoccer/sg/logic_serv" , "copy_cfg":1 , "cmd":"./logic_serv -D"},    
    {"name":"logic_serv-2" , "bin":["./bin/logic_serv/logic_serv"] , "host":"10.161.37.104" , "host_dir":"/home/nmsoccer/sg/logic_serv" , "copy_cfg":1 , "cmd":"./logic_serv -D"},
    {"name":"db_serv-1" ,   "bin":["./bin/db_serv/db_serv"] , "host":"10.144.172.215" , "host_dir":"/home/nmsoccer/sg/db_serv-1" , "copy_cfg":1 , "cmd":"./db_serv"},
    {"name":"db_serv-2" ,   "bin":["./bin/db_serv/db_serv"] , "host":"10.144.172.215" , "host_dir":"/home/nmsoccer/sg/db_serv-2" , "copy_cfg":1 , "cmd":"./db_serv"},
    {"name":"db_serv-3" ,   "bin":["./bin/db_serv/db_serv"] , "host":"10.144.172.215" , "host_dir":"/home/nmsoccer/sg/db_serv-3" , "copy_cfg":1 , "cmd":"./db_serv"}	
  ],

  "proc_cfgs":[
    {"name":"conn_serv-1" ,  "cfg_name":"conf/conn_serv.cfg" , "cfg_tmpl":"./tmpl/conn_serv.tmpl" , "tmpl_param":"id=1001,ip=x.x.x.x,port=10280,name=conn_serv-1"},
    {"name":"conn_serv-2" ,  "cfg_name":"conf/conn_serv.cfg" , "cfg_tmpl":"./tmpl/conn_serv.tmpl", "tmpl_param":"id=1002,ip=x.x.x.x,port=10280,name=conn_serv-2"},
    {"name":"logic_serv-1" , "cfg_name":"logic_serv.cfg" , "cfg_tmpl":"./tmpl/logic_serv.tmpl" , "tmpl_param":"id=2001,name=logic_serv-1"},
	{"name":"logic_serv-2" , "cfg_name":"logic_serv.cfg" , "cfg_tmpl":"./tmpl/logic_serv.tmpl" , "tmpl_param":"id=2002,name=logic_serv-2"},
	{"name":"db_serv-1" ,    "cfg_name":"conf/db/db_serv.cfg" ,    "cfg_tmpl":"./tmpl/db_serv.tmpl" , "tmpl_param":"id=3001"},
	{"name":"db_serv-2" ,    "cfg_name":"conf/db/db_serv.cfg" ,    "cfg_tmpl":"./tmpl/db_serv.tmpl" , "tmpl_param":"id=3002"},
	{"name":"db_serv-3" ,    "cfg_name":"conf/db/db_serv.cfg" ,    "cfg_tmpl":"./tmpl/db_serv.tmpl" , "tmpl_param":"id=3003"}
  ]
   
  }
  ```
上面是一个简单的游戏配置，一共三类进程，conn_serv,logic_serv,db_serv，分别简单代表游戏的接入层逻辑层和数据层进程。conn_serv和logic_serv组成一组，一共两组。一组部署到分发机本机，一组部署到内网另一台机器10.161.37.104上；db_serv部署到内网第三台机器10.144.172.215上，并且预计部署三个实例. 最后，在部署成功之后分别在执行相应应的文件。    
procs选项在上面已经说过了，这里重点介绍proc_cfgs选项：
  * 【name】: **必须**：这里的名字与procs项目填写的name必须对应保持一致，用于标明是哪个进程或者分发的配置文件  
  * 【cfg_name】：**必须**：该进程或者分发生成的独有配置文件名。注意如果包含路径则表示相对路径，最终生成的配置文件会根据cfg_name里的目录层次放置在proc.host_dir目录下  
  * 【cfg_tmpl】: **必须**：如果确定需要为该进程生成配置文件则必须制定该模板路径。模板里如果有参数则以$开头占位。比如conn_serv.tmpl如下所示:
  ```
  cat tmpl/conn_serv.tmpl 
  #PROC ID
  proc_id=$id

  #PROC_NAME
  proc_name="$name"

  #ADDR
  ip=$ip
  listen_port=$port

  #SPACE
  space="test"

  #MAX-CONN
  max_conn=1024

  #TIMEOUT
  timeout=30
  ```
  * 【tmpl_param】：**可选**：如果cfg_tmpl制定的模板文件里有相关的占位符，则tmpl_param需要指定该占位符的值，形式为key=value。不同的组以,分割.比如conn_serv_1的为`{"name":"conn_serv-1" ,  "cfg_name":"conf/conn_serv.cfg" , "cfg_tmpl":"./tmpl/conn_serv.tmpl" , "tmpl_param":"id=1001,ip=x.x.x.x,port=10280,name=conn_serv-1"}`, 那么结合对应的模板文件，最终会生成conn_serv.cfg如下所示：
  ```
  cat cfg/simple_game/conn_serv-1/conf/conn_serv.cfg 
  #PROC ID
  proc_id=1001

  #PROC_NAME
  proc_name="conn_serv-1"

  #ADDR
  ip=x.x.x.x
  listen_port=10280

  #SPACE
  space="test"

  #MAX-CONN
  max_conn=1024

  #TIMEOUT
  timeout=30
  ```
  
### 进阶演示 
这里使用simple_game来进行说明

* 进入demo/simple_game/目录 执行./init.sh进行初始化工作
* 修改./simple_game.json配置文件里的host_dir为本机的有效目录
* 修改./simple_game.json配置文件里的host为有效IP，如果暂无多台机器可以使用配置文件simple_game_local.json部署到本机测试
* 生成各任务配置：  
  ```
  ./spush -C -f ./simple_game.json
  spush starts...
  create cfg...
  create ./cfg/simple_game/conn_serv-1/conf/conn_serv.cfg success!
  create ./cfg/simple_game/conn_serv-2/conf/conn_serv.cfg success!
  create ./cfg/simple_game/logic_serv-1/logic_serv.cfg success!
  create ./cfg/simple_game/logic_serv-2/logic_serv.cfg success!
  create ./cfg/simple_game/db_serv-1/conf/db/db_serv.cfg success!
  create ./cfg/simple_game/db_serv-2/conf/db/db_serv.cfg success!
  create ./cfg/simple_game/db_serv-3/conf/db/db_serv.cfg success!
  
  ```
  生成配置文件成功，并放到了当前的cfg目录中
  
  * 推送任务到各机器:
  ```
  ./spush -P -f demo/simple_game.json 
  spush starts...
  push all procs
  create ./cfg/simple_game/conn_serv-1/conf/conn_serv.cfg success!
  create ./cfg/simple_game/conn_serv-2/conf/conn_serv.cfg success!
  create ./cfg/simple_game/logic_serv-1/logic_serv.cfg success!
  create ./cfg/simple_game/logic_serv-2/logic_serv.cfg success!
  create ./cfg/simple_game/db_serv-1/conf/db/db_serv.cfg success!
  create ./cfg/simple_game/db_serv-2/conf/db/db_serv.cfg success!
  create ./cfg/simple_game/db_serv-3/conf/db/db_serv.cfg success!
  ...
  ----------Push <simple_game> Result---------- 
  ok
  [7/7]
  [db_serv-3]::success 
  [conn_serv-1]::success 
  [conn_serv-2]::success 
  [logic_serv-1]::success 
  [logic_serv-2]::success 
  [db_serv-1]::success 
  [db_serv-2]::success
  ```
  我们看到都已部署成功鸟。
  
  * 验证本机:
    ```
    tree /home/nmsoccer/sg/
    sg/
    |-- conn_serv
    |   |-- conf
    |   |   `-- conn_serv.cfg
    |   |-- conn_serv
    |   |-- conn_serv.cfg
    |   `-- log
    `-- logic_serv
        |-- log
        |-- logic_serv
        `-- logic_serv.cfg
    ```
    因为我们在设置里面有部署后执行对应的脚本文件比如`{"name":"conn_serv-1" , "bin":["./bin/conn_serv/conn_serv"] , "host":"127.0.0.1" , "host_dir":"/home/nmsoccer/sg/conn_serv" , "copy_cfg":1 , "cmd":"./conn_serv -D"},` 会执行./conn_serv -D,而conn_serv是一个简单的脚本文件
    ```
    cat /home/nmsoccer/sg/conn_serv/conn_serv
    #!/bin/bash
    log="./log"
    ts=`date +"%F %T"`
    echo "conn_serv $1 starts at $ts..." >> $log
    ```
    会将当前时间写到本地目录的./log文件里，我们检查下是否完成：
    ```
    cat /home/nmsoccer/sg/conn_serv/log 
    conn_serv -D starts at 2020-03-13 20:03:05...
    ```
    OK,没有问题。
    
  * 验证10.161.37.104：
    这台机器上部署了另外一组conn_serv,logic_serv进程对，它们实例分别为conn_serv-2，logic_serv-2：
    ```
    tree /home/nmsoccer/sg/; cat /home/nmsoccer/sg/logic_serv/logic_serv.cfg 
    /home/nmsoccer/sg/
    |-- conn_serv
    |   |-- conn_serv
    |   |-- conf
    |   |   `-- conn_serv.cfg
    |   `-- log
    `-- logic_serv
        |-- log
        |-- logic_serv
        `-- logic_serv.cfg

    #PROC ID
    proc_id=2002

    #PROC_NAME
    proc_name="logic_serv-2"

    #MAX_PLAYER
    max_player=5000
    ```
    可以确认是正常的
    
  * 验证10.144.172.215：
    这里部署了db_serv的三个实例：
    ```
    tree /home/nmsoccer/sg/
    /home/nmsoccer/sg/
    |-- db_serv-1
    |   |-- conf
    |   |   `-- db
    |   |       `-- db_serv.cfg
    |   |-- db_serv
    |   `-- log
    |-- db_serv-2
    |   |-- conf
    |   |   `-- db
    |   |       `-- db_serv.cfg
    |   |-- db_serv
    |   `-- log
    `-- db_serv-3
        |-- conf
        |   `-- db
        |       `-- db_serv.cfg
        |-- db_serv
        `-- log

    ```
    其中db_serv.cfg保持了配置的目录结构，符合预期
    
  * 选择推送 选择推送logic_serv
  ```
  /spush -f demo/simple_game.json -p "logic*"

  ++++++++++++++++++++spush (2020-03-20 21:08:10)++++++++++++++++++++
  push some procs:logic*
  matched procs num:2
  create cfg:7/7
  ...
  ----------Push <simple_game> Result---------- 
  ok
  [2/2]
  [logic_serv-1]::success 
  [logic_serv-2]::success 
  ```
