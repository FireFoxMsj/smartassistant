// Package task 设备场景任务运行模块；接收，编排，运行场景任务
// task.Manager 启动会加载 scene，包装成 Task，并且加入优先级队列，然后设定每天 23:55:00 进行第二天任务编排
// scene 对应的 Task 运行时，会将对应 scene task 包装成 Task，并且加入优先级队列
package task
