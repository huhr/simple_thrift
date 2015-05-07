#### 场景
当我们在写一些较为简单的程序时，通常使用一种编程语言便可以满足我们的需求。    
比如一些公司内部使用的内容管理系统，或者小的blog。在一些简单的web应用里往     
往使用MVC模式已经可以把逻辑划分的很明确。这些场景下，LAMP基本上搞定了所        
有问题，然而在一些场景下，系统的复杂度较高，业务逻辑较多，性能要求较高，这         
时候系统的分层和模块化势在必行，这时候thrift就派上用场了。      

thrift的使用不仅仅将各种语言黏合在我们的系统中，同时也使得系统的逻辑得以模块      
化，可维护性更高。

#### 以golang为例，看下thrift的构成和实现    

thrift的通信结构分为了transport层，protocol层，processor层以及server层，而其中
processor由定义的接口文件生成，其他层面上不同实现的选择也有很大的灵活性。

##### Transport层:
transport是thrift层次结构中的最底层，Thrift定义了两个transport的接口     
TTransport和TServerTransport。TTransport定义了基本的读写接口以及Flush, Open, IsOpen, Peek    
四个接口,TServerTransport定义了Listen, Accept, Close, Interrupt四个函数。   

StreamTransport:    
	建立在io.Reader和io.Writer之上的transport，StreamTransport分为只读的，只写的和    
读写的三种。thfit提供了StreamTransportFactory来获取实例。典型的应用是积分墙的    
数据导出模块以及AdServer的导入协程的实现。

	func (this *AdImporter) importBonusFromFile(filePath string)(bonus[]*data_shared.Bonus, err error) {
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
回值的实现，会导致client已经关闭了socket但是go server还是继续会往clinet写数据，抛出   
异常，类似：

	Error while flushing write buffer of size 58 to transport, only wrote ...

这个异常便是在FramedTransport中抛出的，即server的返回值已经写到缓存里了Flush的时     
候client已经关闭socket。
	
##### Protocol层:
Protocol层的功能也就是我们说的序列化和反序列化，数据在传输过程中，我们需要定义数据    
的格式，这样才能保证server和client能够识别传输的数据。Protocol是建立在transport之上的，    
thrift提供了json，binary，compact等序列化方式。
