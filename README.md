> 一起学习 k8s 吧  
> 有问题联系：2304262737@qq.com
# k8s-webhook
使用 kubebuilder 为 k8s 原生资源创建 webhook，在 pod 创建之前修改 pod

## 项目创建流程
1. 按照一般 crd 开发方式生成代码框架
    ```shell
    mkdir k8s-webhook
    cd k8s-webhook
    go mod init ydkmm-webhook
    kubebuilder init --domain ydkmm.fun --license none --owner "ydkmm"
    # 根据提示执行 go get 和 go mod tidy
    go get sigs.k8s.io/controller-runtime@v0.12.2
    go mod tidy
    kubebuilder create api --group core --version v1 --kind Pod
    Create Resource [y/n]
    n
    Create Controller [y/n]
    n
    go mod tidy
    ```
2. 生成 webhook 代码
    ```shell
    kubebuilder create webhook --group core --version v1 --kind Pod --defaulting --webhook-version v1
    ```
3. 修改框架代码
   1. 由于没有使用到 crd 资源，所以需要修改 `config/default/kustomization.yaml`，打开 webhook 相关内容，关闭 crd 相关内容
   2. 本次实践是完成对 pod 的修改，因此需要删除 validate 部分，需要修改`config/default/webhookcainjection_patch.yaml`
   3. 对于`config/rbac/kustomization.yaml`，删除 role 相关内容
   4. 手动生成资源清单
   ```shell
    make manifests
    kustomize build config/default/ > k8s-webhook-pod.yaml
    ```
4. 修改逻辑代码
   1. 因为没有使用到 crd 资源，所以不能使用常规的 SetupWebhookWithManager 方法，该方法是对 crd struct 添加的方法，无法对 k8s 原生资源添加该方法，所以修改 `main.go`，删除 SetupWebhookWithManager 方法，注册 webhook
   2. 修改`api/v1/pod_webhook.go`，实现 Handler 方法，使用 admission.decoder 保证资源修改的安全性和版本兼容性
5. 构建镜像并上传仓库，修改`Dokcerfile`
```shell
docker build -t ydkmm/k8s-webhook-pod:v1 .
docker push ydkmm/k8s-webhook-pod:v1
```
6. 修改`k8s-webhook-pod.yaml`，修改 ClusterRoleBinding 使得服务获取集群资源操作权限，并使用刚刚生成的镜像
7. 部署 webhook
```shell
# 先部署 cert-manager，github 下载一个版本
kubectl apply -f cert-manager.yaml

# 利用生成的资源清单一键部署
kubectl apply -f k8s-webhook-pod.yaml

# 查看部署是否成功
kubectl get pod -n k8s-webhook-system
NAME                                              READY   STATUS    RESTARTS   AGE
k8s-webhook-controller-manager-665dfc5b77-rhtlk   2/2     Running   0          2m2s
```
8. 测试功能
```shell
# 部署测试 deploy，查看对应的 pod yaml 文件可以看到多了一个 nginx container
kubectl apply -f deploy-test.yaml

# 查看 pod 可以看到，两个镜像都 ready
kubectl get po -o wide
NAME                           READY   STATUS    RESTARTS   AGE   IP            NODE       NOMINATED NODE   READINESS GATES
test-deploy-68fdc5f679-7vb7r   2/2     Running   0          36s   10.244.0.42   minikube   <none>           <none>
test-deploy-68fdc5f679-ntg97   2/2     Running   0          36s   10.244.0.41   minikube   <none>           <none>
test-deploy-68fdc5f679-rwvwl   2/2     Running   0          36s   10.244.0.43   minikube   <none>           <none>
```
# 相关文档
[准入控制器参考 | Kubernetes](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/)

[What's a webhook? - The Kubebuilder Book](https://book.kubebuilder.io/reference/webhook-overview)

[cert-manager/cert-manager: Automatically provision and manage TLS certificates in Kubernetes (github.com)](https://github.com/cert-manager/cert-manager/)
