
k8s dashboard admin:		
    https://github.com/kubernetes/dashboard

First,you need to install the heapster in the cluster. The guidence of installing the heapster can be found in the following links:
    https://github.com/kubernetes/heapster
    https://github.com/kubernetes/heapster/blob/master/docs/influxdb.md

Then, install the dashboard with metrics information:
    https://github.com/kubernetes/dashboard/wiki/Getting-started
    (Notice: Every time you want to rebuild and up the dashboard, you need to first delete the "node_modules" directory. And then reinstall the dependencies by the command "npm i --unsafe-perm")



installing step by step:

go:
tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz
cat << EOF > /root/.bash_profile
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin
export GOPATH=$HOME/gopath
EOF
source /root/.bash_profile


nvm:
curl -o- https://raw.githubusercontent.com/creationix/nvm/v0.33.8/install.sh | bash
-------------------------

cat << EOF > /root/.bash_profile
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion
EOF
source /root/.bash_profile
nvm install node
nvm use node


jdk:
tar -C /usr/local -xzf jdk-8u131-linux-x64.tar.gz
cat << EOF > /root/.bash_profile
export JAVA_HOME=/usr/local/jdk1.8.0_131
export PATH=$PATH:$JAVA_HOME/bin
EOF
source /root/.bash_profile
java -version


gulp:
npm install --global gulp-cli
npm install --global gulp


git:
yum install git


base:
yum install make g++ gcc gcc-c++


patch:
yum install -y patch




k8s:
bash
copy /opt/dashboard/k8s
(scp -rf /lwh/dashboard root@10.141.211.171:/opt/dashboard/k8s)
cd /opt/dashboard/k8s

yum install dos2unix -y
chmod a+x build/postinstall.sh
dos2unix build/postinstall.sh
npm i --unsafe-perm

proxy:
kubectl proxy --port=8080
(kill pid-proxy:  ps aux | grep kubectl)
http://10.141.211.171:8080/api/

run:
gulp serve
-------------------------
1. Install the required Babel packages: npm install gulp-babel babel-preset-es2015 --save-dev
2. Add a .babelrc file to the root of your project with these contents:
{
  "presets": ["es2015"],
  "compact": true
})
-------------------------
npm rebuild node-sass

url:
http://10.141.211.171:9090/
http://10.141.211.171:3001/



istio:
https://github.com/istio/istio/blob/master/DEV-GUIDE.md



