## MyRaft
---
#### 2021/2/21 星期日
1. 准备将start函数代码分为三个部分编写（分别对应Follower,Candidate,Leader），在Server.go中新写了一个loop函数，并在Server.go的Start函数中调用loop函数，今日将Follower部分写好了。但是发现在投票的时候需要考虑当前任期内是否投过票，所以需要在Server中再加一个变量表示投票的任期用于解决这个问题（哪个任期投了票就把这个变量的值设为这个任期，投票时看投票任期是否与当前任期一致，一致则代表当前任期已经投过票，不一致则可以投票，初始投票任期为0），该部分今日已完成。
2. 在写Follower部分代码时发现当节点需要主动发送数据请求时则需要利用通道技术，可以循环的取通道的值，一旦通道中出现值则就发起数据请求，而通道中放入值则就需要根据具体情况进行编写（比如说一个新的节点添加对等点请求则就在告诉入口节点并同意加入集群之后就往通道中加入值）。
3. 关于对等点加入与删除问题：一个新启动的节点在启动后会主动发起添加对等点请求，在关闭之前会主动发起请求告诉集群中节点删除对等点信息。

+ FollowerLoop函数中未完成工作：添加对等点请求
+ Server的Stop函数中未完成工作：删除对等点请求
---
#### 2021/2/22 星期一

1. CandidateLoop代码完成
2. CandidateLoop中还需要考虑投票的时间（因为当两个以上候选者平分投票，则谁也无法获得超过集群一半的票数，则会不停的等待，当超过投票的时间，设为Follower状态）
3. CandidateLoop中若收到领导者的心跳信息则应该主动设为Follower状态

+ FollowerLoop、CandidateLoop、LeaderLoop中都需要考虑服务器突然下线的情况，以及下线需要做哪些处理？
+ FollowerLoop、CandidateLoop中若收到心跳中的日志索引小于自身的索引则是否需要将日志与领导者保持一致（网络分区）？
---
#### 2021/2/23 星期二

1. LeaderLoop代码完成
2. Follower、Candidater、Leader中都得考虑接收到投票的请求，代码中通过一个vote()函数解决
3. 客户端发送给Leader的添加日志请求先写为key,value形式，这样后期如果需要修改也好修改

+ 客户端添加日志请求通过广播形式发给集群中节点，到底是所有节点（所有节点都接收则如果不是leader需要转发这个添加请求给leader）还是只有Leader接收呢

---

#### 2021/2/24 星期三

1. 在添加对等点的时候，如果集群中节点收到添加对等点请求后自己也发送一个添加对等点请求给发送这个请求的节点，这样两者就都能添加对等点了
2. 添加对等点时，如果是leader接收到这个请求，则需要告诉新加入的节点立即开始心跳，如果是candidater接收到这个请求，则需要告诉新加入的节点立即开始投票

+ 入口节点，也可以称为元服务器（功能：所有节点的信息可以查看，可以管理节点的上线与下线，负责新节点的加入）是否应该独立于集群之外or也是集群中的一个节点（但是不参与集群运转）？
+ 集群中节点下线存在两种情况：一种是入口节点向集群中节点发起下线请求（主动），另一种是节点由于断电、程序出错等方式突然下线（被动）。第一种方式下线比较容易解决，第二种方式下线该如何让集群中其他节点知道这个节点已经下线了呢？

---

#### 2021/2/25 星期四

1. 集群中节点主动下线时，需要考虑删除对等点和关闭日志
2. 当Follower投票后应该将Follower的CurrentTerm+1，这样一旦当前任期无法选出Leader，Follower在超时后会CurrentTerm+1后发起新的一轮投票，就不会造成长时间无法选出Leader的情况

+ 集群中节点主动下线时，除了考虑删除对等点和关闭日志，是否还需要针对不同节点的状态进行不同的关闭操作（Candidate的投票不会再继续收集，Leader的心跳也不会再发送，好像也不太需要其他操作了，可能是我想多了）？
+ 接下来工作的难点：入口节点（元服务器）如何完成代码编写，集群中节点被动下线情况如何解决，日志部分的存储与利用，快照和状态机的实现

---

#### 2021/2/26 星期五

1. 入口节点部分元素和方法设计（不全，以后会陆续补充）

   ```go
   type entrance struct {
   	id              string              //集群的标识
   	ip 				string				//ip
   	recPort 		int					//接收udp消息地址
   	mutex      		sync.RWMutex		//读写锁
   	context     	string				//详细信息
   	currentLeader   string				//当前集群领导者
   	currentTerm     uint64				//当前集群任期
   	peer            map[string]string   //集群节点的信息（名称：ip）
   	pLen            uint64				//集群大小（节点数量）
   }
   
   type Entrance interface {
   	Id()    			string
   	Ip()    			string
   	SetIp(string)
   	RecPort() 			int
   	Context() 			string
   	CurrentLeader() 	string
   	CurrentTerm()   	uint64
   	Peer()    			map[string]string
   	PeerLen() 			uint64
   	Start()
   }
   ```

2. 关于入口节点的一些思考：

   > - 当有Server被New之后想加入集群并启动，必须通过入口节点的允许才可以实现，这个节点发送udp请求给入口节	点，取得入口节点的允许（此时入口节点存储该节点相关信息），入口节点返回一个udp消息，收到后节点才启动并向集群中节点广播添加对等点消息。
   > - 入口节点接收来自集群中Leader的心跳信息，通过心跳信息就可以知道当前集群中的任期和领导者
   > - 入口节点可以管理节点让其主动下线，此时只需要广播即可让集群中节点删除对等点，当集群中有节点被动下线入口节点也需要知道，如何才能实现还未考虑清楚？
   > - 入口节点也可能会产生网络分区的情况，这种情况下该如何解决还未考虑清楚
   > - 入口节点想要知道当前集群中某个节点的详细的信息，由于入口节点存储节点ip，则可以发送udp请求，让该节点返回一个数据包给入口节点
   > - 入口节点是否需要存储集群中所有节点的相关配置信息，如果需要如何存储？

---

#### 2021/2/27 星期六

1. 入口节点加入新节点部分已经完成

2. 加入新节点具体代码思路：新节点在New过之后运行初始化Init函数，在该函数中首先向入口节点发送一个加入集群的udp请求（请求中有新节点接收的地址，端口号为默认初始化端口，不参与集群的广播），入口节点接收后首先判断新节点的名称和ip是否已经存在在集群中，若存在则拒绝加入，否则将通过入口节点的决定后向新节点发送udp加入反馈包，新节点收到请求后判断请求是否通过，若通过则将新节点设为初始化状态，等待Start。

3. 加入请求和响应部分代码块如下：

   ```go
   type joinRequest struct {
      name 	string
      sip   	string
      recPort 	int
      ip 		string
      port 	int
   }
   
   type joinResponse struct {
   	join   bool
   	name   string
   	ip     string
   	port   int
   }
   ```

+ 客户端的请求一般都得先发送给入口节点，然后通过入口节点发送给集群中的leader，入口节点负责监听客户端发来的请求（比如查看集群中节点状态，关闭集群节点，添加集群节点，存值，取值等等）

---

#### 2021/2/28 星期日

1. 入口节点接收来自客户端的请求部分代码完成

   AddKeyOrder（添加key/value）：该请求暂定为发送给leader，后期如果测试丢包则采用广播形式

   StopServer（关闭指定的节点）：让入口节点负责转发这个请求给需要关闭的节点，再通过这个节点广播删除对等点的操作后下线

   DelPeerOrder（删除指定对等点）：一旦有节点下线，则需要广播集群中所有节点（也包含入口节点），入口节点则需要将对应的对等点信息删除

2. 由于可能出现相同的key被修改的情况，则针对每次的添加key的请求都设置一个标识（标识采用不断递增的uint64位变量），可以用于将来的版本控制（查询旧的键值对）

+ 当客户端想要获取集群中存取的某一个值时，应该通过入口节点向Leader取值 or 集群中任意一个节点取值 ？

---

#### 2021/3/1 星期一

1. 今天写代码测试时发现json结构体其中的变量必须大写，这个bug头都大了，查了一下代码，发现很多地方定义的变量都是小写，所以把项目中所有用于发送数据的结构体变量全都改为了大写，用变量的地方也都修改完毕

2. 接收udp的数据包时，发现字节后面自动添加了很多\x00空字符，需要在接收数据时将数据后面多余的\x00去掉后再json解组，不然会出现解组错误

3. 客户端的功能构思：

   > + 查看所有节点的基本信息（显示在一个页面），通过获取所有入口节点的对等点信息即可实现
   > + 查看单个节点的当前状态的详细信息，通过入口节点向集群中对应节点发送数据请求，同时该节点将当前状态的详细信息返还给入口节点，入口节点再返还给客户端
   > + 存储Key/Value信息至集群中，通过入口节点向Leader发送添加日志请求，Leader将对应的请求转换成日志索引后，Leader发送心跳，集群中其他节点收到心跳发现自己的日志索引小于Leader的日志索引，就会向Leader发送一条新的日志添加请求，然后Leader再将这条新的日志通过udp发送给这些节点，从而完成集群中所有节点信息的一致性
   > + 取Key/Value值信息，通过入口节点向Leader发送取值请求，Leader查询自己的日志，将对应的Key/Value值获取后返回给入口节点，入口节点再返还给客户端
   > + 让集群中的节点强制下线，通过入口节点向指定的节点发送下线请求，该节点在完成下线前的数据处理后立即下线
   > + 管理新节点加入集群，一旦有新节点加入集群，则这个节点在Init的时候就向客户端发送进入请求，客户端向入口节点发送查询请求，查看当前节点是否加入集群，若加入集群就不允许再次加入，否则客户端具有决定该节点是否能加入集群的权限
   > + 让一整个集群下线，即让入口节点下线即可，一旦入口节点下线则无法完成客户端与集群的交互，所以理论上就已经处于下线状态了

+ 客户端的所有请求都通过入口节点才能访问到集群内部信息，也就要保证客户端能与入口节点一直相连，一旦该连接断了，则客户端将收不到集群的信息
+ 入口节点与集群的连接十分关键，一旦入口节点下线，则集群中的所有节点的运行状态将无法得知，集群将变得不可用