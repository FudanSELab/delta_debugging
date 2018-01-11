


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
