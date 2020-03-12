# SPush
A simple push tool,一款简单的分发工具.该工具主要目标是将不同的文件或目录分发到不同的机器及对应的目录。 

### 主要特点
* **多重分发**： 支持一次分发多个任务到多台机器及目录  
* **配置简单**： 使用一个简单的json格式的配置文件进行即可使用
* **较少依赖**： 尽量减少对环境的特殊依赖，除了运行分发机器需要部署go环境以外，目标机器只需要普通的unix/linux开发环境即可
* **结果反馈**： 为了保证分发的正确性，所有的分发任务均有分发结果反馈告知是否成功或发生错误
* **配置生成**： 在较大型的比如游戏项目中需要为不同进程编写对应的配置文件，这里支持根据模板通过进程名自动生成相关的配置文件
* **指令运行**： 为了方便一些部署工作，支持针对每个特定的分发过程在分发结束后执行简单的指令  
  _想到了再写_  
  
### 安装
 * 依赖
   * 在分发机器上需要安装golang  
   * 在分发机器上需要能执行expect命令用于同步异地机器
   
 * 安装
   * 下载spush到本地并解压
   * 进入spush主目录./init.sh初始化
   * go build spush.go
   * ./spush xxxx 即可使用

### 命令
  * `-C`: 仅仅为分发任务生成各分发任务自有的配置文件(如果有设置)
  * `-P`: 推送并执行配置里所有定义的分发任务。这里会先执行-C选项的内容
  * `-p xxx`: 根据任务名选择执行配置里某些分发任务
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
   {"name":"cpy1" , "bin":["./tools/" , "./init.sh"] , "host":"" , "host_dir":"/home/nmsoccer/spush_demo/simple_copy/cpy1" , "copy_cfg":0 , "cmd":""} , 
   {"name":"cpy2" , "bin":["/etc/passwd"] , "host":"127.0.0.1" , "host_dir":"/home/nmsoccer/spush_demo/simple_copy/cpy2" , "copy_cfg":0 , "cmd":""} 
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
  * 【cmd】: **可选**：在该分发任务部署成功之后可以选择执行的命令。注意该命令的执行目录是在host_dir的下一层目录。比如设置../xx.sh则会执行到host_dir/xx.sh。
  
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
  ./spush -P -f ./simple_copy.json 
  spush starts...
  push all procs
  .
  ----------Push <simple_copy> Result---------- 
  ok
  [cpy1]::success 
  [cpy2]::success
  ```
  推送成功，这里可以显示两项任务都OK鸟。我们可以去看一看目标的目录是否O杰把K
  ```
  tree /home/nmsoccer/spush_demo/
  /home/nmsoccer/spush_demo/
  `-- simple_copy
      |-- cpy1
      |   |-- init.sh
      |   `-- tools
      |       |-- exe_cmd.exp
      |       |-- peer_exe.sh
      |       |-- push.sh
      |       |-- report.c
      |       `-- scp.exp
      `-- cpy2
          `-- passwd
  ```
  看了下都是OK的了,当然这里已经把痕迹都删除了，如果想要查看部署后的遗留可以加上-r选项
  
  * 保存现场：
  ```
  ./spush -r -P -f ./simple_copy.json 
  spush starts...
  push all procs
  .
  ----------Push <simple_copy> Result---------- 
  ok
  [cpy1]::success 
  [cpy2]::success 
  
  ```
    部署如下：
  ```
  tree /home/leiming/spush_demo/
  /home/leiming/spush_demo/
  `-- simple_copy
      |-- cpy1
      |   |-- cpy1.tar.gz
      |   |-- cpy1.tar.gz.md5
      |   |-- init.sh
      |   |-- spush_simple_copy_tools
      |   |   |-- peer_exe.sh
      |   |   |-- report
      |   |   `-- report.c
      |   `-- tools
      |       |-- exe_cmd.exp
      |       |-- peer_exe.sh
      |       |-- push.sh
      |       |-- report.c
      |       `-- scp.exp
      `-- cpy2
          |-- cpy2.tar.gz
          |-- cpy2.tar.gz.md5
          |-- passwd
          `-- spush_simple_copy_tools
              |-- peer_exe.sh
              |-- report
              `-- report.c
  ```
  我们可以看到部署目录遗留了spush_$task_tools的遗留目录
  
 * 查看日志 可以在/tmp/spush/$task/$proc/log里检查到在部署机器上执行的情况 如下：
   ```
   tree /tmp/spush/
   /tmp/spush/
   `-- simple_copy
       |-- cpy1
       |   `-- log
       `-- cpy2
           `-- log
   ```
   都是这样的层次
