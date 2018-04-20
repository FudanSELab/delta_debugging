package helper;


public class LoginResult {

    private boolean status;

    private String message;

    private Account account;

    private String token;

    private int loginNum;

    public LoginResult(){
        //Default Constructor
    }


    public int getLoginNum() {
        return loginNum;
    }

    public void setLoginNum(int loginNum) {
        this.loginNum = loginNum;
    }

    public boolean getStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public Account getAccount() {
        return account;
    }

    public void setAccount(Account account) {
        this.account = account;
    }

    public String getToken() {
        return token;
    }

    public void setToken(String token) {
        this.token = token;
    }
}
