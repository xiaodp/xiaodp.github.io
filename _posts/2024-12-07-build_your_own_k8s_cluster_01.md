---
layout: post
title: "搭建你自己的k8s集群-01"
date:   2024-12-07
tags: [k8s]
comments: true
author: xiaodp
toc: true
---

使用VirtualBox + Centos 搭建自己的k8s集群
<!-- more -->

# 【1】搭建k8s集群

# 1 环境准备

k8s搭建方式：

kind
minikube
虚拟机

采用虚拟机部署，更加接近真实场景。

## 1.1 主机准备

- 3台centos7 虚拟机
    - ip列表
    
    ```yaml
    192.168.33.10
    192.168.33.11
    192.168.33.12
    ```
    
    - 规格：2C4G
    - 节点互相ping通，能访问外网

## 1.2 集群规划

- k8s-node1: 192.168.33.10
- k8s-node2: 192.168.33.11
- k8s-node3: 192.168.33.12
- 设置主机名
    - hostnamectl set-hostname k8s-node1
    - hostnamectl set-hostname k8s-node2
    - hostnamectl set-hostname k8s-node3
- 更改host文件
    
    ```bash
    192.168.33.10 k8s-node1
    192.168.33.11 k8s-node2
    192.168.33.12 k8s-node3
    
    cat >>  /etc/hosts <<EOF
    192.168.33.10 k8s-node1
    192.168.33.11 k8s-node2
    192.168.33.12 k8s-node3
    EOF
    ```
    

# 2 安装步骤

## 2.1 关闭防火墙

```bash
systemctl stop firewalld && systemctl disable firewalld
```

## 2.2 关闭iptables

```sql
stop iptables && systemctl disable iptables
```

## 2.3 关闭SELINUX

```sql
sestatus
sed -i s#SELINUX=enforcing#SELINUX=disabled# /etc/selinux/config
reboot
```

## 2.4 关闭交换分区

```bash
swapoff -a && sed -ri 's/.swap.*/#&/' /etc/fstab
```

## 2.5 yum更换国内源

```sql
mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.backup

curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo

yum makecache
```

## 2.6 同步服务时间

```bash
yum install -y ntpdate
```

## 2.7 配置网络

```bash

cat > /etc/modules-load.d/containerd.conf << EOF
overlay
br_netfilter
EOF

cat > /etc/sysctl.d/k8s.conf << EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
EOF

modprobe overlay
modprobe br_netfilter

sysctl -p /etc/sysctl.d/k8s.conf
```

- 配置ipvs
    
    ```sql
    # 1.安装ipset和ipvsadm
    [root@master ~]# yum install ipset ipvsadm -y
    # 2.添加需要加载的模块写入脚本文件
    [root@master ~]# cat <<EOF> /etc/sysconfig/modules/ipvs.modules
    #!/bin/bash
    modprobe -- ip_vs
    modprobe -- ip_vs_rr
    modprobe -- ip_vs_wrr
    modprobe -- ip_vs_sh
    modprobe -- nf_conntrack_ipv4
    EOF
    # 3.为脚本添加执行权限
    [root@master ~]# chmod +x /etc/sysconfig/modules/ipvs.modules
    # 4.执行脚本文件
    [root@master ~]# /bin/bash /etc/sysconfig/modules/ipvs.modules
    # 5.查看对应的模块是否加载成功
    [root@master ~]# lsmod | grep -e ip_vs -e nf_conntrack_ipv4
    
    ```
    

## 2.8 安装containerd

```bash
yum install -y yum-utils device-mapper-persistent-data lvm2

yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo

yum install -y containerd.io cri-tools

cat >  /etc/containerd/config.toml <<EOF
disabled_plugins = ["restart"]
[plugins.linux]
shim_debug = true
[plugins.cri]
sandbox_image = "registry.aliyuncs.com/google_containers/pause:3.9"
EOF

systemctl enable containerd && systemctl start containerd && systemctl status containerd
```

## 2.9 重启

## 2.10 添加k8s源

```bash
cat>>/etc/yum.repos.d/kubernetes.repo<<EOF
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
```

## 2.11 安装k8s

```bash
yum install -y kubelet kubeadm kubectl

systemctl enable kubelet && systemctl start kubelet && systemctl status kubelet
```

指定k8s的cgroup配置与网络配置

```sql
echo 'KUBELET_CGROUP_ARGS="--cgroup-driver=systemd"' >> /etc/sysconfig/kubelet
echo 'KUBE_PROXY_MODE="ipvs"' >> /etc/sysconfig/kubelet
```

- master节点初始化集群

master节点上执行kubeadm init， 其他节点上执行kubeadm join

[https://www.cnblogs.com/lmgsanm/p/16221470.html](https://www.cnblogs.com/lmgsanm/p/16221470.html)

```bash
kubeadm init \
--apiserver-advertise-address=192.168.33.10 \
--pod-network-cidr=10.244.0.0/16 \
--service-cidr=10.96.0.0/12 \
--image-repository=registry.aliyuncs.com/google_containers

Your Kubernetes control-plane has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

Alternatively, if you are the root user, you can run:

  export KUBECONFIG=/etc/kubernetes/admin.conf

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

## 在worker节点上执行：
**kubeadm join 192.168.56.10:6443 --token mwis5z.59f5r78nmzrvwncq \
	--discovery-token-ca-cert-hash sha256:166e9fcc2f30f4e731df3df008094c0e4019287681b9978443ef334cff3b9bed**
```

- 如果报错，使用来清理

```sql
kubeadm reset
```

```sql

kubeadm join 192.168.33.10:6443 --token x29qha.os02rji0f0jqwlgm \
	--discovery-token-ca-cert-hash sha256:081777f584b009d2c1f286126110e993c5294763bb61528d2d9e476e9cfacb1d
```

- start using your cluster

```bash
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

- plane节点执行

```bash

kubectl get nodes

# 查看命名空间的pod
kubectl get pods -A

watch -n 1 -d kubectl get nodes
```

- 加入节点

```bash
kubeadm join 192.168.56.10:6443 --token mwis5z.59f5r78nmzrvwncq --discovery-token-ca-cert-hash sha256:166e9fcc2f30f4e731df3df008094c0e4019287681b9978443ef334cff3b9bed
```

- 查看节点状态

```bash
[vagrant@k8s-node1 ~]$ kubectl get nodes
NAME        STATUS     ROLES           AGE    VERSION
k8s-node1   NotReady   control-plane   19h    v1.27.3
k8s-node2   NotReady   <none>          105s   v1.27.3
k8s-node3   NotReady   <none>          108s   v1.27.3
```

- 配置containerd代理

```bash
mkdir /etc/systemd/system/containerd.service.d/

cat>>/etc/systemd/system/containerd.service.d/http-proxy.conf<<EOF
[Service]
Environment="HTTP_PROXY=http://192.168.0.104:7890"
Environment="HTTPS_PROXY=http://192.168.0.104:7890"
Environment="NO_PROXY=localhost,127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,.svc,.cluster.local,.ewhisper.cn,<nodeCIDR>,<APIServerInternalURL>,<serviceNetworkCIDRs>,<etcdDiscoveryDomain>,<clusterNetworkCIDRs>,<platformSpecific>,<REST_OF_CUSTOM_EXCEPTIONS>"
EOF

systemctl daemon-reload && systemctl restart containerd
systemctl status containerd
```

NO_PROXY一定配置，否则kubelet会尝试通过代理访问，从而报错

- 配置集群网络
    - calico
        
        ```sql
        wget https://calico-v3-25.netlify.app/archive/v3.25/manifests/calico.yaml
        export https_proxy=http://192.168.0.104:7890 http_proxy=http://192.168.0.104:7890 all_proxy=socks5://192.168.0.104:7890
        kubectl apply -f calico.yaml
        
        在各个节点上手动pull景象
        ctr -n k8s.io images pull docker.io/calico/cni:v3.25.0
        ctr -n k8s.io images pull docker.io/calico/node:v3.25.0
        ctr -n k8s.io images pull docker.io/calico/kube-controllers:v3.25.0
        docker.io/library/nginx:1.14-alpine
        ```
        
    - flannel
        
        ```bash
        kubectl apply -f kube-flannel.yml
        ```
        
        - kube-flannel.yml
            
            ```bash
            ---
            kind: Namespace
            apiVersion: v1
            metadata:
              name: kube-flannel
              labels:
                k8s-app: flannel
                pod-security.kubernetes.io/enforce: privileged
            ---
            kind: ClusterRole
            apiVersion: rbac.authorization.k8s.io/v1
            metadata:
              name: flannel
            rules:
            - apiGroups:
              - ""
              resources:
              - pods
              verbs:
              - get
            - apiGroups:
              - ""
              resources:
              - nodes
              verbs:
              - get
              - list
              - watch
            - apiGroups:
              - ""
              resources:
              - nodes/status
              verbs:
              - patch
            ---
            kind: ClusterRoleBinding
            apiVersion: rbac.authorization.k8s.io/v1
            metadata:
              name: flannel
            roleRef:
              apiGroup: rbac.authorization.k8s.io
              kind: ClusterRole
              name: flannel
            subjects:
            - kind: ServiceAccount
              name: flannel
              namespace: kube-flannel
            ---
            apiVersion: v1
            kind: ServiceAccount
            metadata:
              name: flannel
              namespace: kube-flannel
            ---
            kind: ConfigMap
            apiVersion: v1
            metadata:
              name: kube-flannel-cfg
              namespace: kube-flannel
              labels:
                tier: node
                app: flannel
            data:
              cni-conf.json: |
                {
                  "name": "cbr0",
                  "cniVersion": "0.3.1",
                  "plugins": [
                    {
                      "type": "flannel",
                      "delegate": {
                        "hairpinMode": true,
                        "isDefaultGateway": true
                      }
                    },
                    {
                      "type": "portmap",
                      "capabilities": {
                        "portMappings": true
                      }
                    }
                  ]
                }
              net-conf.json: |
                {
                  "Network": "10.244.0.0/16",
                  "Backend": {
                    "Type": "vxlan"
                  }
                }
            ---
            apiVersion: apps/v1
            kind: DaemonSet
            metadata:
              name: kube-flannel-ds
              namespace: kube-flannel
              labels:
                tier: node
                app: flannel
            spec:
              selector:
                matchLabels:
                  app: flannel
              template:
                metadata:
                  labels:
                    tier: node
                    app: flannel
                spec:
                  affinity:
                    nodeAffinity:
                      requiredDuringSchedulingIgnoredDuringExecution:
                        nodeSelectorTerms:
                        - matchExpressions:
                          - key: kubernetes.io/os
                            operator: In
                            values:
                            - linux
                  hostNetwork: true
                  priorityClassName: system-node-critical
                  tolerations:
                  - operator: Exists
                    effect: NoSchedule
                  serviceAccountName: flannel
                  initContainers:
                  - name: install-cni-plugin
                    #image: docker.io/flannel/flannel-cni-plugin:v1.1.2
                    image: docker.io/rancher/mirrored-flannelcni-flannel-cni-plugin:v1.1.0
                    command:
                    - cp
                    args:
                    - -f
                    - /flannel
                    - /opt/cni/bin/flannel
                    volumeMounts:
                    - name: cni-plugin
                      mountPath: /opt/cni/bin
                  - name: install-cni
                    #image: docker.io/flannel/flannel:v0.22.0
                    image: docker.io/rancher/mirrored-flannelcni-flannel:v0.20.2
                    command:
                    - cp
                    args:
                    - -f
                    - /etc/kube-flannel/cni-conf.json
                    - /etc/cni/net.d/10-flannel.conflist
                    volumeMounts:
                    - name: cni
                      mountPath: /etc/cni/net.d
                    - name: flannel-cfg
                      mountPath: /etc/kube-flannel/
                  containers:
                  - name: kube-flannel
                    #image: docker.io/flannel/flannel:v0.22.0
                    image: docker.io/rancher/mirrored-flannelcni-flannel:v0.20.0
                    command:
                    - /opt/bin/flanneld
                    args:
                    - --ip-masq
                    - --kube-subnet-mgr
                    resources:
                      requests:
                        cpu: "100m"
                        memory: "50Mi"
                    securityContext:
                      privileged: false
                      capabilities:
                        add: ["NET_ADMIN", "NET_RAW"]
                    env:
                    - name: POD_NAME
                      valueFrom:
                        fieldRef:
                          fieldPath: metadata.name
                    - name: POD_NAMESPACE
                      valueFrom:
                        fieldRef:
                          fieldPath: metadata.namespace
                    - name: EVENT_QUEUE_DEPTH
                      value: "5000"
                    volumeMounts:
                    - name: run
                      mountPath: /run/flannel
                    - name: flannel-cfg
                      mountPath: /etc/kube-flannel/
                    - name: xtables-lock
                      mountPath: /run/xtables.lock
                  volumes:
                  - name: run
                    hostPath:
                      path: /run/flannel
                  - name: cni-plugin
                    hostPath:
                      path: /opt/cni/bin
                  - name: cni
                    hostPath:
                      path: /etc/cni/net.d
                  - name: flannel-cfg
                    configMap:
                      name: kube-flannel-cfg
                  - name: xtables-lock
                    hostPath:
                      path: /run/xtables.lock
                      type: FileOrCreate
            ```
            
- 查看pods状态，直到全部为running

```bash
[vagrant@k8s-node1 ~]$ kubectl get pods -A
NAMESPACE              NAME                                         READY   STATUS    RESTARTS   AGE
kube-flannel           kube-flannel-ds-dczqc                        1/1     Running   0          11m
kube-flannel           kube-flannel-ds-sg2hl                        1/1     Running   0          19m
kube-flannel           kube-flannel-ds-zlnkh                        1/1     Running   0          6m1s
kube-system            coredns-7bdc4cb885-ddzzl                     1/1     Running   0          21h
kube-system            coredns-7bdc4cb885-x6nt4                     1/1     Running   0          21h
kube-system            etcd-k8s-node1                               1/1     Running   0          21h
kube-system            kube-apiserver-k8s-node1                     1/1     Running   0          21h
kube-system            kube-controller-manager-k8s-node1            1/1     Running   0          21h
kube-system            kube-proxy-9x4v6                             1/1     Running   0          21h
kube-system            kube-proxy-ch7xv                             1/1     Running   0          110m
kube-system            kube-proxy-drrw8                             1/1     Running   0          110m
kube-system            kube-scheduler-k8s-node1                     1/1     Running   0          21h
kubernetes-dashboard   dashboard-metrics-scraper-764cf47594-ttkbj   1/1     Running   0          25m
kubernetes-dashboard   kubernetes-dashboard-68997bf576-rd6nb        1/1     Running   0          25m
```

- 添加storage class
    
    [https://github.com/rancher/local-path-provisioner](https://github.com/rancher/local-path-provisioner)
    
    ```bash
    
    kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.30/deploy/local-path-storage.yaml
    
    kubectl get sc
    
    [root@k8s-node1 linux-amd64]# kubectl get pod -n local-path-storage
    NAME                                      READY   STATUS    RESTARTS   AGE
    local-path-provisioner-65d5864f8d-qbtkf   1/1     Running   0          13m
    
    # 要设置为default, 才能在没有指定sc的情况下参与制备
    kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
    ```