# 设备场景
对智汀家庭云即smart-assistant（以下简称SA）的设备场景的说明。
## 场景
场景是指通过SA实现设备联动。例如，自动检测今天的天气情况，今天无雨，定时智能音箱播放浇花提醒，并且播报今天的天气情况。
根据自身需求，把多种控制并发的事情编辑成一个场景，并命名，可以通过场景控制很多设备，实现一键操作的功能。

## 场景的相关操作

### 创建场景
创建智能场景前请确保您的家庭已添加设备，且用户是否拥有创建场景的权限。

#### 场景名称
场景名称在该家庭下需要确保唯一性。

#### 触发条件
通过配置触发条件，达到条件后能执行对应的任务，并且可以设置触发条件的生效时段。触发条件分为三种
* 手动执行，点击即可执行
* 定时执行，如每天8点
* 设备状态变化时，如开灯时，感应到人时

当触发条件为手动触发时只能添加一种触发条件。而选择其他两种可以添加多种，同时需要确定条件关系。条件关系可以选择
* 满足所有条件
* 满足任一条件

##### 技术实现
系统中启动一个服务，作为消息队列（以下简称smq）的消费者，消费者不断去轮训消息队列，看看有没有新的数据，如果有就消费。
查看下面为伪代码：
```
for {
    select {
    case ct := <-ticker.C:
        fmt.Printf("current ticket at: %d:%d \n", ct.Minute(), ct.Second())
        if pq.Len() == 0 {
            ticker.Reset(sleepTickTime)
            continue
        }

        task := heap.Pop(pq).(*Task)
        now := time.Now()
        timeAt := now.Unix()

        if task.Priority > timeAt {
            nextTick := time.Unix(task.Priority, 0).Sub(now)
            ticker.Reset(nextTick)
            qs.push(task)
        } else {
            ticker.Reset(defaultTickTime)
            go task.Run()
        }
    }
}
```
当设置为手动执行的场景时，会添加一条任务数据，执行时间为当前时间，加进smq，等待消费者消费。
```
t = NewTask(WrapSceneFunc(scene, false), 0)
PushTask(t, scene)
```
而设置为自动执行的场景时，计算任务今天的下次执行时间，并添加任务数据，加进smq，等待消费者消费。
```
// 获取任务今天的下次执行时间
days := time.Now().Sub(c.TimingAt).Hours() / 24
nextTime := c.TimingAt.AddDate(0, 0, int(days))
t = NewTaskAt(WrapSceneFunc(scene, true), nextTime)
PushTask(t, scene)
```
如果自动执行场景的生效时段为重复性，那么会在每天 23:55:00 进行第二天任务编排
```
// AddArrangeSceneTask 每天定时编排场景任务
func AddArrangeSceneTask(executeTime time.Time) {
    var f TaskFunc
    f = func(task *Task) error {
        addSceneTaskByTime(executeTime.AddDate(0, 0, 1))
    
        // 将下一个定时编排任务排进队列
        AddArrangeSceneTask(executeTime.AddDate(0, 0, 1))
        return nil
    }
    
    task := NewTaskAt(f, executeTime)
    MinHeapQueue.Push(task)
}

// 每天 23:55:00 进行第二天任务编排
AddArrangeSceneTask(now.EndOfDay().Add(-5 * time.Minute))
```

#### 执行任务
当满足触发条件后，可以自动执行配置好的执行任务。执行任务认为两种
* 智能设备，如开灯，播放音乐
* 控制场景，如开启夏季晚会场景

##### 技术实现
任务执行，通过消费者消费smq中的任务，去执行run方法去执行对应的任务。
```
func (item *Task) Run() {
    fmt.Println("Run ", item.ToString())
    if item.f != nil {
        f := item.f
        for _, wrapper := range item.wrappers {
            f = wrapper(f)
        }
        if err := f(item); err != nil {
            log.Println("task run err:", err)
        }
    }
}
```


### 查看场景
场景分成 “手动” 和 “自动” 两个执行类型，页面加载时判断用户是否拥有控制场景的权限，在页面展示中 “手动”场景排在“自动”场景的上方；

* 手动类场景为“执行”按键，可直接点击触发执行任务
* 自动类场景为“开关”按键，设置打开或者关闭状态


## 注意事项
* 场景的修改和控制不仅仅取决于用户是否拥有修改和控制场景的权限，还包括该用户是否有对场景中的设备操作项的控制权限。
    * eg：如果您拥有控制场景A的权限，但是您没有场景A里面设备B的开关控制权限，则您同样没有控制该场景A的权限。修改场景也是如此。