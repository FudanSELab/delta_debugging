package apiserver.util;

public class Cluster {
    private String name;
    private String masterIp;
    private String username;
    private String passwd;
    private String token;
    private String apiServer;

    public Cluster(){

    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getMasterIp() {
        return masterIp;
    }

    public void setMasterIp(String masterIp) {
        this.masterIp = masterIp;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getPasswd() {
        return passwd;
    }

    public void setPasswd(String passwd) {
        this.passwd = passwd;
    }

    public String getApiServer() {
        return apiServer;
    }

    public void setApiServer(String apiServer) {
        this.apiServer = apiServer;
    }

    public String getToken() {
        return token;
    }

    public void setToken(String token) {
        this.token = token;
    }

}
