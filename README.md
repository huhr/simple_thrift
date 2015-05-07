#### 场景
当我们在写一些较为简单的程序时，通常使用一种编程语言便可以满足我们的需求。    
比如一些公司内部使用的内容管理系统，或者小的blog。在一些简单的web应用里往    
往使用MVC模式已经可以把逻辑划分的很明确。这些场景下，LAMP基本上搞定了所    
有问题，然而在一些场景下，系统的复杂度较高，业务逻辑较多，性能要求较高，这    
时候系统的分层和模块化势在必行，这时候thrift就派上用场了。    

thrift的使用不仅仅将各种语言黏合在我们的系统中，同时也使得系统的逻辑得以模块    
化，可维护性更高。    

#### 以golang为例，看下thrift的构成和实现    

thrift的通信结构分为了transport层，protocol层，processor层以及server层，    
而其中processor由定义的接口文件生成，其他层面上不同实现的选择也有很大的    
灵活性。

##### Transport层:
transport是thrift层次结构中的最底层，Thrift定义了两个transport的接口    
TTransport和TServerTransport。TTransport定义了基本的读写接口以及    
Flush, Open, IsOpen, Peek四个接口,TServerTransport定义了Listen,    
Accept, Close, Interrupt四个函数。   

TStreamTransport:    
建立在io.Reader和io.Writer之上的transport，StreamTransport分为只读的，只写的和    
读写的三种。thfit提供了StreamTransportFactory来获取实例。典型的应用是积分墙的    
数据导出模块以及AdServer的导入协程的实现。    

	func (this *AdImporter) importBonusFromFile(filePath string) (bonus[]*data_shared.Bonus, err error) {
		bonus = make([]*data_shared.Bonus, 0)
		fd, err := os.Open(filePath)
		defer fd.Close()
		if err == nil {
			transport := thrift.NewStreamTransportR(fd)
			protocol := thrift.NewTBinaryProtocolTransport(transport)
			for {
				b := aow_adserver_types.NewAowExtraBonus()
				err = b.Read(protocol)
				if err != nil {
					break
				 } 
				 bonus = append(bonus, data_shared.NewBonus(b))
			} 
		} 
		return bonus, nil 
	}

TSocket:    
TSocket也是实现了TTransport定义的接口，带有duck type特性的语言不易一眼就看出    
来继承关系。TSocket读写操作都在网络连接net.Conn上进行。其实TStreamTransport    
和TSocket才直接进行数据流读写操作，后面的两种transport都是通过组合的方式，在    
流操作的基础上做的封装。    

TServerSocket:    
TServerSocket是实现了TServerTransport，用来监听端口，每次获取到新的连接，交给    
上层处理。

BufferedTransport:    
增加了读写缓存，在进行IO操作时，频繁的进行read/write系统调用会消耗更多的资源，    
使用缓存的好处在于每次进行读写操作时都能一次性的读写更多的数据，减少系统调用    
次数，从而提升性能。

FramedTransport:    
在组合TTransport的基础之上，增加了1M读写缓存的transport，这里是定义了一套自    
己的传输协议按帧传输。每一帧的前四个byte表示这一帧包含的数据的byte数，每次先    
读前四个byte，然后将特定大小的一帧数据读入自身的缓存中，再做处理，写的时候也    
是先写入缓存中，Flush时计算缓存中数据长度，再发送出去。线上生产环境golang的    
server都是使用FramedTransport来进行通信的。    

thrift 0.9.1版本对one way的接口生成的golang代码有一个bug，server端生成的代码在处   
理过请求之后，会继续向cilent发送处理结果，由于php的client并没有接收one way函数返    
回值的实现，会导致client已经关闭了socket但是go server还是继续会往clinet写数据，    
抛出异常，类似：

	Error while flushing write buffer of size 58 to transport, only wrote ...

这个异常便是在FramedTransport中抛出的，即server的返回值已经写到缓存里了Flush的时    
候client已经关闭socket。

##### Protocol层:
Protocol层的功能也就是我们说的序列化和反序列化，数据在传输过程中，我们需要定义数据    
的格式，这样才能保证server和client能够识别传输的数据。Protocol是建立在transport之    
上的，thrift提供了json，binary，compact等序列化方式。    

##### Processor层:
processor做的就是读入数据，处理数据，再把处理的结果写出去。接口只定义了一个Process    
函数。processor的实现由thrift的接口定义文件生成，我们在启动server的时候要给processor    
传handler参数，这个handler便是我们自己写的对定义接口的实现。

##### Server层:    
server层显然就是在server端使用，由他来负责整体的调度，协调各层的使用接口定义很简    
单，只有启动的Serve()接口和Stop()接口。go版本只有一个TSimpleServer，较py简单很多，    
依赖于go的语言优势。    
server有6个属性，processor工厂，TServerTransport，分别负责读和写的transport，    
protocol工厂。而我们看到的形如NewTSimpleServer4，NewTSimpleServer6都只是做了    
一层封装，让server端能够使用不同的transport和protocol。其中NewTSimpleServer2    
就是将TSocket作为输入输出transport，TBinaryProtocol作为输入输出的protocol；    
NewTSimpleServer4的transport，protocol取决于我们的传入参数，但是会保证server    
的输入输transport，protocol是一致的；NewTSimpleServer6最灵活，我们可以分别控    
制server的输入输出transport和protocol。这样读写数据使用的协议是不同的。    

需要扩展一下为什么py的server各种各样，go却只有一种。先看Server()的实现:

	func (p *TSimpleServer) Serve() error {
		p.stopped = false
		err := p.serverTransport.Listen()
		if err != nil {
			return err
		}
		for !p.stopped {
			client, err := p.serverTransport.Accept() 
			if err != nil {
				log.Println("Accept err: ", err)
			}
			if client != nil {
				go func() {  
					if err := p.processRequest(client); err != nil {
						log.Println("error processing request:", err)
					}
				}()
			}
		}
	return nil
	}

我们可能会注意到这个实现，go server在启动以后，是一直监听着端口，每次Accept到    
了新的连接，就把连接交给一个协程去处理。首先，处理是一个异步的过程，不同连接    
的处理是没有关联的，这样当请求量非常大的时候，server端可能会有成千上万个协程    
在工作，幸运的是协程要比线程更加容易驾驭，所以go server可以胜任很大的并发量级。    
对比py的TProcessPoolServer，其实是由固定数量的线程，处理请求的过程其实是同步的，    
如果全部的workers都在处理请求中，client发的新的请求就被阻塞住了。    

那么问题来了，为什么goroutine更容易驾驭。    

	while self.isRunning.value:
		try:
			client = self.serverTransport.accept()
			self.serveClient(client)
		except (KeyboardInterrupt, SystemExit):
			return 0
		except Exception, x:
			logging.exception(x)

还是先看TProcessPoolServer的实现，每一个worker的工作内容都是一样的，这里进行IO操    
作时，总会或多或少的遇到等待，IO的速度跟CPU的速度差了很多，例如在read的时候阻塞住    
了，这时候这个worker就开始等待了那么这个等待的时间里，这个线程其实是闲置的，没有    
进行任何计算。生产环境中，我看可以pstree看到py的server有很多线程，但其实这些线程    
工作并不是饱和的。为了解决这个问题，TNonblockingServer出现了，这是一个非阻塞IO的    
server实现，在这个实现中，workers并没有拿到socket连接进行处理，而是主进程管理所有    
的连接，通过系统调用select，拿到所有不用等待的连接，放在数组中，分配给workers处理。    
让workers在请求量大的时候都有活干，而不会把时间浪费在等待某一个连接的IO上。   

goroutine是构建在线程之上的概念，是由go自身来进行调度的，这里在实现的时候使用epoll    
模型，当某一个协程要等待IO的时候，go会调度执行该goroutine的线程去执行其他的goroutine，    
当IO就绪的时候，再回调激活阻塞的goroutine，调度空闲的线程来执行，这样go server的线程    
总会选择可以执行的goroutine执行，工作就饱和了许多。不管是系统调度还是goroutine的调度，    
一定是有资源消耗的，只是比操作系统调度线程，进程更加轻量级。    
