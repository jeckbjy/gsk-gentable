# gsk-gentable

generate csv,tsv,json and code for golang from xlsx sheet

## 功能特性

- 数据源:  
  - 支持本地xlsx
  - 支持google doc(gdoc)  
    连接需要认证,使用命令 "gentable auth"会读取credentials.json配置文件并生成所需要的配置信息
  - 支持指定某些sheet
- 数据输出:
  - 支持csv,tsv
  - json,value全部都是字符类型
- 代码输出:
  - 支持golang,可指定Package名字
- 表头格式:
  - 第一行名字,第二行类型,第三行注释
- 支持相同列名合并成一个数组,
  - 规则:{名字}_{索引[0...n]},限制是数值类型
- 支持的类型:SID,NID,INT,UINT,INT64,UINT64,FLOAT32,FLOAT64,BOOL,LIST,MAP,DATE,WEEK,ENUM
  - 类型限制:SID,NID代表字符ID和数值ID(INT),必须是第一列,指定该类型则会自动生成map索引,否则不会建立索引,  
  - LIST,MAP的KEY,VALUE只能是数值类型
  - DATE格式必须为 yyyy-mm-dd HH:MM:SS 如 2006-01-02 15:04:05,时间可以不全填写,不填写则补0
  - WEEK格式必须为 WEEK-HH:MM:SS 如5-15:04:05,星期可以不填写,表示每天这个时间点

## 输出目录结构

例如:输出csv,json数据,go代码  
output  
 ┣ csv  
 ┃ ┣ data  
 ┃ ┗ go  
 ┗ json  
 ┃ ┣ data  
 ┃ ┗ go  

## TODO

- 更多语言的支持
- 详细的日志信息
- LIST添加Filter功能,用于累加该列数据或者新定义一种类型(RLIST,RandList),该功能通常用于概率随机