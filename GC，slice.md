### 1. go的GC过程

[原文链接](https://blog.csdn.net/weixin_39998006/article/details/100928939)

GC的优点：

* 可以将未被任何对象引用的对象进行回收，避免悬挂指针
* 只有回收器可以释放对象，不会出现二次释放
* 回收器掌握堆中对象的全局信息以及可能访问堆中对象的线程信息，可以决定任意对象是否需要回收
* 回收器管理对象，模块之间减少耦合

Go的GC：

* 基于Mark Sweep，不过是并发的Mark和并发的Sweep，即并发GC（两层含义）

  * 每个mark或sweep本身是多个线程(协程)执行的——concurrent

    * GC时整体进行STW，对象引用关系不再改变，对mark或sweep进行分块，就能多个线程(协程)执行任务mark或sweep

  * mutator(应用程序)和collector(收集器)同时运行——background

    * 实现相对复杂，因为mutator会改变已经被scan对象的引用关系，示例：
      ```go
      b.obj1=c
       
                              gc mark start
       
                              gc scan a
       
      mutaotr  a.obj1=c
       
      mutator  b.obj1=nil
       
                              gc scan b
       
                              gc mark termination
       
                              sweep and free c(error)
      b有c的引用. gc开始, 先扫描了a, 然后mutator运行, a引用了c, b不再引用c, gc再扫描b, 
      然后sweep, 清除了c. 这里其实a还引用了c, 导致了正确性问题.
      ```

    * 引入写屏障，是在写入指针前执行的一小段代码用于防止指针丢失. 这一小段代码Golang是在编译时写入的. Golang目前写屏障在mark阶段开启

    * 将c的指针写入到a.obj1之前, 会先执行一段判断代码, 如果c已经被扫描过, 就不再扫描, 如果c没有被扫描过, 就把c加入到待扫描的队列中. 这样就不会出现丢失存活对象的问题存在

  * 三色标记法：并发的GC算法，原理：

    ```
    1. 首先创建三个集合：白, 灰, 黑. 白色节点表示未被mark和scan的对象, 灰色节点表示已经被mark, 但是还没有scan的对象, 而黑色表示已经mark和scan完的对象.
    2. 初始时所有对象都在白色集合.
    3. 从根节点开始广度遍历, 将其引用的对象加入灰色集合.
    4. 遍历灰色集合, 将灰色对象引用的白色对象放入灰色集合, 之后将此灰色对象放入黑色集合.
    
    在Go Runtime的实现中, 并没有白色集合, 灰色集合, 黑色集合这样的容器：
    白色对象: 某个对象对应的gcMarkBit为0(未被标记)
    灰色对象: gcMarkBit为1(已被标记)且在(待scan)gcWork的待scan buffer中
    黑色对象: gcMarkBit为1(已被标记)且不在(已经scan)gcWork的待scan buffer中
    
    分代GC：GC后的gcMarkBits不清空, 对象存活为1. 那下一次GC时, 在还没有进行标记时, 发现gcMarkBits为1, 那就是老对象, 为0, 就是新分配的对象
    ```

    ![](https://imgconvert.csdnimg.cn/aHR0cHM6Ly91cGxvYWQtaW1hZ2VzLmppYW5zaHUuaW8vdXBsb2FkX2ltYWdlcy82NzgzNTY1LWI3MjgyY2UzZDM4NzJiOGYuZ2lmP2ltYWdlTW9ncjIvYXV0by1vcmllbnQvc3RyaXB8aW1hZ2VWaWV3Mi8yL3cvNDMwL2Zvcm1hdC93ZWJw)





### 2.slice的底层原理

[原文链接](https://blog.csdn.net/lengyuezuixue/article/details/81197691)

Go的数组是值类型，赋值和函数传参操作都会赋值整个数组的数据

* 切片是对数组一个连续片段的引用，是一个引用类型。这个片段可以是整个数组，或者是由起始和终止索引标识的一些项的子集。

* ```go
  type slice struct {  
      array unsafe.Pointer
      len   int
      cap   int
  }
  ```

  ![](https://img-blog.csdn.net/20180725103740874?watermark/2/text/aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2xlbmd5dWV6dWl4dWU=/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70)

  Pointer是指向一个数组的指针，len代表当前切片的长度，cap是容量

  ![](https://img-blog.csdn.net/20180725103816476?watermark/2/text/aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2xlbmd5dWV6dWl4dWU=/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70)

* 创建切片要用make函数
* nil切片与空切片：区别在于，空切片指向的地址不是nil，指向的是一个内存地址，但是它没有分配任何内存空间，即底层元素包含0个元素
* 扩容策略：
  * 如果新容量大于旧容量的两倍，则直接采用新容量；
  * 如果新容量小于等于旧容量：
    * 如果旧切片的长度小于1024，那么新的容量等于两倍的旧容量
    * 如果旧切片的长度大于等于1024，那么通过for循环每次增加1/4的容量直到大于等于预期容量
* 建议用字面量创建切片的时候，cap 的值一定要保持清醒，避免共享原数组导致的 bug
* 原来数组的容量已经达到了最大值，再想扩容， Go 默认会先开一片内存区域，把原来的值拷贝过来，然后再执行 append() 操作。这种情况丝毫不影响原数组
* 切片拷贝：copy函数
  * slicecopy ：slicecopy 方法会把源切片值(即 fm Slice )中的元素复制到目标切片(即 to Slice )中，并返回被复制的元素个数，copy 的两个类型必须一致。slicecopy 方法最终的复制结果取决于较短的那个切片，当较短的切片复制完成，整个复制过程就全部完成了
  * slicestringcopy：字节数组拷贝
* 如果用range遍历切片，拿到的value是切片中的值拷贝，所以每次打印value的地址都不变，需要通过&slice[index]获取地址































