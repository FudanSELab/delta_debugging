package apiserver.util;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.UnsupportedEncodingException;

import ch.ethz.ssh2.*;
import org.apache.commons.lang.StringUtils;

public class RemoteExecuteCommand {

    //字符编码默认是utf-8
    private static String  DEFAULTCHART="UTF-8";
    private Connection conn;
    private String ip;
    private String userName;
    private String userPwd;

    public RemoteExecuteCommand() {
        //do nothing
    }

    public RemoteExecuteCommand(String ip, String userName, String userPwd) {
        this.ip = ip;
        this.userName = userName;
        this.userPwd = userPwd;
    }

    /**
     * 远程登录linux的主机
     * @author Ickes
     * @since  V0.1
     * @return
     *      登录成功返回true，否则返回false
     */
    public Boolean login(){
        boolean flg = false;
        try {
            conn = new Connection(ip);
            conn.connect();//连接
            flg = conn.authenticateWithPassword(userName, userPwd);//认证
        } catch (IOException e) {
            e.printStackTrace();
        }
        return flg;
    }

    /**
     * @author Ickes
     * 远程执行shll脚本或者命令
     * @param cmd
     *      即将执行的命令
     * @return
     *      命令执行完后返回的结果值
     * @since V0.1
     */
    public String execute(String cmd){
        String result="";
        try {
            if(login()){
                Session session = conn.openSession();//打开一个会话
                session.execCommand(cmd);//执行命令
                result = processStdout(session.getStdout(),DEFAULTCHART);
                //如果为得到标准输出为空，说明脚本执行出错了
                if(StringUtils.isBlank(result)){
                    result = processStdout(session.getStderr(),DEFAULTCHART);
                }
                conn.close();
                session.close();
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
        return result;
    }


    /**
     * @author Ickes
     * 远程执行shll脚本或者命令
     * @param cmd
     *      即将执行的命令
     * @return
     *      命令执行成功后返回的结果值，如果命令执行失败，返回空字符串，不是null
     * @since V0.1
     */
    public String executeSuccess(String cmd){
        String result="";
        try {
            if(login()){
                Session session= conn.openSession();//打开一个会话
                session.execCommand(cmd);//执行命令
                result = processStdout(session.getStdout(),DEFAULTCHART);
                conn.close();
                session.close();
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
        return result;
    }

    /**
     * 解析脚本执行返回的结果集
     * @author Ickes
     * @param in 输入流对象
     * @param charset 编码
     * @since V0.1
     * @return
     *       以纯文本的格式返回
     */
    private String processStdout(InputStream in, String charset){
        InputStream stdout = new StreamGobbler(in);
        StringBuffer buffer = new StringBuffer();
        try {
            BufferedReader br = new BufferedReader(new InputStreamReader(stdout,charset));
            String line;
            while((line = br.readLine()) != null){
                buffer.append(line+"\n");
            }
        } catch (UnsupportedEncodingException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
        return buffer.toString();
    }

    /**
     * 上传本地文件到远程服务器端，即将本地的文件localFile上传到远程Linux服务器中的配置的目录下
     * @param localFile
     */
    public void uploadFile(String localFile) {
        System.out.println("[=====] 开始上传文件:" + localFile + "");
        try{
            if(login()) {
                SCPClient scpClient = conn.createSCPClient();
                scpClient.put(localFile, 10000,"./","0644");
                System.out.println("[=====] 上传文件结束:" + localFile + "");
            }
        }catch(IOException e){
            e.printStackTrace();
        }
        conn.close();
    }

    public void modifyFile(String fileName, String svcName) {
        System.out.println("[=====] 开始修改文件:" + fileName + "");
        try{
            if(login()) {
                //先删除文件，再传入数据
                rmFile(fileName);
                String data = fillLongRuleFile(svcName);

                login();
                SFTPv3Client client = new SFTPv3Client(conn);
                SFTPv3FileHandle handle = client.createFile(fileName);

                byte []arr = data.getBytes();
                client.write(handle, 0, arr, 0, arr.length);
                client.closeFile(handle);
                client.close();
                System.out.println("[=====] 修改文件完毕:" + fileName + "");
            }
        }catch(IOException e){
            e.printStackTrace();
        }
        conn.close();
    }

    public void modifyFileWithSourceSvcName(String fileName, String svcName, String srcSvcName) {
        System.out.println("[=====] 开始修改文件[With Source Svc]:" + fileName + "");
        try{
            if(login()) {
                //先删除文件，再传入数据
                rmFile(fileName);

                System.out.println("modifyFileWithSourceSvcName:  " + svcName + " = " + srcSvcName);
                String data = fillLongRuleFileWithSource(svcName,srcSvcName);

                login();
                SFTPv3Client client = new SFTPv3Client(conn);
                SFTPv3FileHandle handle = client.createFile(fileName);

                byte []arr = data.getBytes();
                client.write(handle, 0, arr, 0, arr.length);
                client.closeFile(handle);
                client.close();
                System.out.println("[=====] 修改文件完毕[With Source Svc]:" + fileName + "");
            }
        }catch(IOException e){
            e.printStackTrace();
        }
        conn.close();
    }

    public String fillLongRuleFile(String svcName){
        String longRuleStr = "apiVersion: config.istio.io/v1alpha2\n" +
                "kind: RouteRule\n" +
                "metadata:\n" +
                "  name: rest-service-delay-long-svcName\n" +
                "spec:\n" +
                "  destination:\n" +
                "    name: svcName\n" +
                "  httpFault:\n" +
                "    delay:\n" +
                "      percent: 100\n" +
                "      fixedDelay: 10000s\n" +
                " \n";
        String fullLongRuleString = longRuleStr.replaceAll("svcName",svcName);
        System.out.println("[=====]fullLongRuleString");
        System.out.println(fullLongRuleString);
        return fullLongRuleString;
    }

    public String fillLongRuleFileWithSource(String svcName, String sourceSvcName){
        String longRuleStr = "apiVersion: config.istio.io/v1alpha2\n" +
                "kind: RouteRule\n" +
                "metadata:\n" +
                "  name: rest-service-delay-long-svcName\n" +
                "spec:\n" +
                "  destination:\n" +
                "    name: svcName\n" +
                "  match:\n" +
                "    source:\n" +
                "      name: sourceSvcName \n" +
                "    request:\n" +
                "      headers:\n" +
                "        cookie: \n" +
                "          regex: \"^(.*?;)?(jichao=dododo)(;.*)?$\" \n" +
                "  httpFault:\n" +
                "    delay:\n" +
                "      percent: 100\n" +
                "      fixedDelay: 10000s\n" +
                " \n";
        String fullLongRuleString = longRuleStr.replaceAll("svcName",svcName);
        String fullLongRuleStringFinal = fullLongRuleString.replaceAll("sourceSvcName",sourceSvcName);
        System.out.println("[=====]fullLongRuleString");
        System.out.println(fullLongRuleStringFinal);
        return fullLongRuleStringFinal;
    }

    /**
     * 删除远端Linux服务器上的文件
     * @param filePath
     */
    public void rmFile(String filePath) {
        System.out.println("[=====] 开始删除文件:" + filePath + "");
        try{
            if(login()) {
                SFTPv3Client sftpClient = new SFTPv3Client(conn);
                sftpClient.rm(filePath);
                System.out.println("[=====] 删除文件完毕:" + filePath + "");
            }
        }catch(IOException e){
            e.printStackTrace();
        }
        conn.close();
    }

    /**
     * 连接和认证远程Linux主机
     * @return boolean
     */

    public static void setCharset(String charset) {
        DEFAULTCHART = charset;
    }

    public Connection getConn() {
        return conn;
    }

    public void setConn(Connection conn) {
        this.conn = conn;
    }

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public String getUserName() {
        return userName;
    }

    public void setUserName(String userName) {
        this.userName = userName;
    }

    public String getUserPwd() {
        return userPwd;
    }

    public void setUserPwd(String userPwd) {
        this.userPwd = userPwd;
    }

}
