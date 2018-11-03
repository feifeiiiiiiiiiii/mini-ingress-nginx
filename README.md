# Mini-Ingress-Nginx
A simple ingress-nginx implements, let me to study k8s and ingress related  knowledge, base `Nginx Inc kubernetes-ingress`,
detail doc you can look up [nginxinc/kubernetes-ingress](https://github.com/nginxinc/kubernetes-ingress)

```
  _   _      _ _         __  __ _       _   ___                                 _   _       _
 | | | | ___| | | ___   |  \/  (_)_ __ (_) |_ _|_ __   __ _ _ __ ___  ___ ___  | \ | | __ _(_)_ __ __  __
 | |_| |/ _ \ | |/ _ \  | |\/| | | '_ \| |  | || '_ \ / _` | '__/ _ \/ __/ __| |  \| |/ _` | | '_ \\ \/ /
 |  _  |  __/ | | (_) | | |  | | | | | | |  | || | | | (_| | | |  __/\__ \__ \ | |\  | (_| | | | | |>  <
 |_| |_|\___|_|_|\___/  |_|  |_|_|_| |_|_| |___|_| |_|\__, |_|  \___||___/___/ |_| \_|\__, |_|_| |_/_/\_\
                                                      |___/                           |___/
```

# Steps


1. Deploy Nginx Ingress

```

kubectl create -f deployments/namespace.yaml

kubectl create -f deployments/nginx-ingress-dep.yaml

kubectl create -f deployments/nginx-ingress-svc.yaml


```

2. Setup Ingress example


```

kubectl create -f ./ingress-example

```

3. Edit /etc/hosts


```

127.0.0.1 nsq.proxy.com

```

4. Use `curl` test

```

curl http://nsq.proxy.com:${NodePort}/proxy/info

input eg:
  {"version":"1.1.0","broadcast_address":"nsq-broker-86c486d6-dxwbm","hostname":"nsq-broker-86c486d6-dxwbm","http_port":4151,"tcp_port":4150,"start_time":1541219672}

```

5. Good luck have fun

# Nginx Ingress logs

```
2018/11/03 12:20:46 Hello Mini Ingress Nginx
2018/11/03 12:20:46 Writing NGINX conf to /etc/nginx/nginx.conf
2018/11/03 12:20:46 The main NGINX config file has been updated
2018/11/03 12:20:46 Starting nginx
2018/11/03 12:20:46 [notice] 11#11: using the "epoll" event method
2018/11/03 12:20:46 [notice] 11#11: nginx/1.15.5
2018/11/03 12:20:46 [notice] 11#11: built by gcc 6.3.0 20170516 (Debian 6.3.0-18+deb9u1)
2018/11/03 12:20:46 [notice] 11#11: OS: Linux 4.9.93-linuxkit-aufs
2018/11/03 12:20:46 [notice] 11#11: getrlimit(RLIMIT_NOFILE): 1048576:1048576
2018/11/03 12:20:46 [notice] 11#11: start worker processes
2018/11/03 12:20:46 [notice] 11#11: start worker process 15
2018/11/03 12:20:46 Adding endpoints: nsq
2018/11/03 12:20:46 Adding Ingress: nsq-ingress
2018/11/03 12:20:46 Adding or Updating Ingress: mini-nginx-ingress/nsq-ingress
2018/11/03 12:20:46 Error getting service nsq: service mini-nginx-ingress/nsq doesn't exist
2018/11/03 12:20:46 Error retrieving endpoints for the service nsq: service mini-nginx-ingress/nsq doesn't exist
2018/11/03 12:20:46 Adding service: nsq
2018/11/03 12:20:46 Reloading nginx. configVersion
2018/11/03 12:20:46 [notice] 20#20: signal process started
2018/11/03 12:20:46 [notice] 11#11: signal 1 (SIGHUP) received from 20, reconfiguring
2018/11/03 12:20:46 [notice] 11#11: reconfiguring
2018/11/03 12:20:46 [notice] 11#11: using the "epoll" event method
2018/11/03 12:20:46 [notice] 11#11: start worker processes
2018/11/03 12:20:46 [notice] 11#11: start worker process 21
2018/11/03 12:20:46 AddedOrUpdated Configuration for mini-nginx-ingress/nsq-ingress was added or updated
2018/11/03 12:20:46 Adding or Updating Ingress: mini-nginx-ingress/nsq-ingress
2018/11/03 12:20:46 Reloading nginx. configVersion
2018/11/03 12:20:46 [notice] 23#23: signal process started
2018/11/03 12:20:46 AddedOrUpdated Configuration for mini-nginx-ingress/nsq-ingress was added or updated
2018/11/03 12:20:46 [notice] 11#11: signal 1 (SIGHUP) received from 23, reconfiguring
2018/11/03 12:20:46 [notice] 11#11: reconfiguring
2018/11/03 12:20:46 [notice] 15#15: gracefully shutting down
2018/11/03 12:20:46 [notice] 15#15: exiting
2018/11/03 12:20:46 [notice] 15#15: exit
2018/11/03 12:20:46 [notice] 11#11: using the "epoll" event method
2018/11/03 12:20:46 [notice] 11#11: start worker processes
2018/11/03 12:20:46 [notice] 11#11: start worker process 24
2018/11/03 12:20:46 [notice] 21#21: gracefully shutting down
2018/11/03 12:20:46 [notice] 21#21: exiting
2018/11/03 12:20:46 [notice] 21#21: exit
2018/11/03 12:20:46 [notice] 11#11: signal 17 (SIGCHLD) received from 15
2018/11/03 12:20:46 [notice] 11#11: worker process 15 exited with code 0
2018/11/03 12:20:46 [notice] 11#11: signal 29 (SIGIO) received
2018/11/03 12:20:46 [notice] 11#11: signal 17 (SIGCHLD) received from 21
2018/11/03 12:20:46 [notice] 11#11: worker process 21 exited with code 0
2018/11/03 12:20:46 [notice] 11#11: signal 29 (SIGIO) received
2018/11/03 12:34:52 Adding service: nginx-ingress
2018/11/03 12:34:52 Adding endpoints: nginx-ingress
192.168.65.3 - - [03/Nov/2018:12:36:46 +0000] "GET /proxy/info HTTP/1.1" 200 163 "-" "curl/7.54.0" "-"
2018/11/03 12:36:46 [info] 24#24: *10 client 192.168.65.3 closed keepalive connection
192.168.65.3 - - [03/Nov/2018:12:36:48 +0000] "GET /proxy/info HTTP/1.1" 200 163 "-" "curl/7.54.0" "-"
2018/11/03 12:36:48 [info] 24#24: *12 client 192.168.65.3 closed keepalive connection
192.168.65.3 - - [03/Nov/2018:12:36:49 +0000] "GET /proxy/info HTTP/1.1" 200 163 "-" "curl/7.54.0" "-"
2018/11/03 12:36:49 [info] 24#24: *14 client 192.168.65.3 closed keepalive connection
192.168.65.3 - - [03/Nov/2018:12:36:54 +0000] "GET /proxy/info HTTP/1.1" 200 163 "-" "curl/7.54.0" "-"
2018/11/03 12:36:54 [info] 24#24: *16 client 192.168.65.3 closed keepalive connection
192.168.65.3 - - [03/Nov/2018:12:36:57 +0000] "GET /proxy/info HTTP/1.1" 200 163 "-" "curl/7.54.0" "-"
2018/11/03 12:36:57 [info] 24#24: *18 client 192.168.65.3 closed keepalive connection
192.168.65.3 - - [03/Nov/2018:12:48:57 +0000] "GET /proxy/info HTTP/1.1" 200 163 "-" "curl/7.54.0" "-"
2018/11/03 12:48:57 [info] 24#24: *20 client 192.168.65.3 closed keepalive connection
```