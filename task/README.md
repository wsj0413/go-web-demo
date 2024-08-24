获取所有任务 ：

```shell
curl http://localhost:8080/api/tasks
```

#### 创建任务：

```shell
curl -X POST -H "Content-Type: application/json" -d '{"title": "Sample Task"}' http://localhost:8080/api/tasks
```

#### 更新任务状态 ：

```shell
curl -X PUT -H "Content-Type: application/json" -d '{"done": true}' http://localhost:8080/api/tasks/1
```

#### 删除任务：

```shell
curl -X DELETE http://localhost:8080/api/tasks/1
```
