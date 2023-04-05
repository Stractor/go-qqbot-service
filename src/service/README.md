## botAction
管理bot的服务注册
在配置文件里的CmdList中，配置name、func和desc，然后在本文件夹中增加("BotAction_%s",func)这个方法就能在启动服务的时候，使用name作为关键字运行命令

!!!!!! func不能包含中文 !!!!!!

因为目前是用来做qq机器人的，所以运行命令的返回都是字符串。所有错误都以字符串的形式返回