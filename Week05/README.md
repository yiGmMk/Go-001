1. 可用性
- 可用性挑战
   (1)如何在流量突增、依赖服务宕机等外界紧急情况发生时不需要人工干预来自动做到快速止损、防止整个分布式系统雪崩
   基本抗灾和容错能力:流控、熔断,服务降级和快速恢复 
   (2)


2. 参考 Hystrix 实现一个滑动窗口计数器
- 雪崩
  用户重试 + 代码逻辑重试

- 应对 

- Hystrix(限流,熔断): 资源隔离 + 熔断器  + 命令模式
- 资源隔离
  为每个服务分配独立的资源(线程池等),资源隔离
- 熔断器
  根据服务健康状况(请求失败数/总数)打开或关闭熔断器开关:
  1. 开启:  请求禁止通过,经过一段时间自动进入半开状态,此时只允许少量 请求通过=>请求成功则关闭开关
  2. 关闭:  closed（关闭状态，流量可以正常进入）
  3. half-open: (半开状态，open状态持续一段时间后将自动进入该状态，重新接收流量，一旦请求失败，重新进入open状态，但如果成功数量达到阈值，将进入closed状态)

- 命令模式
  在命令模式中添加有服务调用失败后的降级逻辑
- 降级策略
  1. 两种降级模型，即信号量（同步）模型和线程池（异步）模型
  2. 信号量
  3. 线程池

- 断路器
  1. 滑动窗口
   断路器需要的时间窗口请求量和错误率这两个统计数据，都是指固定时间长度内的统计数据，断路器的目标，就是根据这些统计数据来预判并决定系统下一步的行为，Hystrix通过滑动窗口来对数据进行“平滑”统计，默认情况下，一个滑动窗口包含10个桶（Bucket），每个桶时间宽度是1秒，负责1秒的数据统计
  2. 每个桶都记录着1秒内的四个指标数据：成功量、失败量、超时量和拒绝量，这里的拒绝量指的就是【信号量/线程池资源检查】中被拒绝的流量。10个桶合起来是一个完整的滑动窗口，所以计算一个滑动窗口的总数据需要将10个桶的数据加起来。
  3.  

- 参考:
  1. https://blog.csdn.net/manzhizhen/article/details/79591132
  2. https://blog.csdn.net/manzhizhen/article/details/80296655
  3. 