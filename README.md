# go-qqbot-service

## 对gpt失去兴趣了。。。下面没有了

### 配合qq机器人使用的服务端

## 功能列表
- [ ] 通过wss收发qq消息（包括语音，语音直接转文字发gpt）
  - [x] 文字信息收发
- [ ] 通过http收发qq消息（待定）
- [x] 转发消息到chatGPT
  - [x] 设置人设
  - [x] 设置chatGPT参数（如记忆时间等）
  - [x] 人设管理、重新生成会话、清空会话
  - [x] 临时人设（无预设新会话）
  - [ ] 增加管理员
  - [ ] 增加隐藏模式功能
  - [ ] 群聊模式（每人一个会话、共享会话）
- [ ] 识图
- [ ] 图片生成（暂不考虑）
  - [ ] 图生图
  - [ ] 文生图
  - [ ] 修改大模型和lora
- [ ] 语音条生成（等学会tts）
- [ ] 搞录播
- [ ] 支持跑团
#### 环境
- 语言：go

#### 运行指导
目前是基于go-cqhttp，先运行go-cqhttp再运行这个程序，否则看起来会闪退（目前没做重连，下次想起来再做吧）
go-cqhttp选择wss服务端，ip写127.0.0.1，端口号写这个程序配置的端口号。两个都配通就能跑起来

我增加了tasks，如果使用vsc的task，直接运行就能在output目录下找到exe文件。
把exe文件和两个配置文件config.yml/user_config.yml放到相同文件夹下运行即可

#### 遇到的问题
  - 我在访问openai的api时，明明已经用v2ray科学上网了，但是在vsc里和cmd里就是连不上。后来搜了一下，这篇文章帮助了我。
    - https://blog.csdn.net/SHERLOCKSALVATORE/article/details/123599042
    - 成功之后，vsc也没配上（我只是在cmd里set http_proxy），没有修改环境变量，就用了这篇文章的方法，修改了vsc的代理
    - https://ost.51cto.com/answer/5159