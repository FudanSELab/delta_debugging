package cluster7;

import org.openqa.selenium.By;
import org.openqa.selenium.JavascriptExecutor;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.remote.DesiredCapabilities;
import org.openqa.selenium.remote.RemoteWebDriver;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

import java.net.URL;
import java.util.concurrent.TimeUnit;

public class TestLoginErrorInstance {
    private WebDriver driver;
    private String trainType;//0--all,1--GaoTie,2--others
    private String baseUrl;
    private int loginNumber;
    public static void login(WebDriver driver,String username,String password){
        ((JavascriptExecutor) driver).executeScript("arguments[0].scrollIntoView();",  driver.findElement(By.id("flow_one_page")));
        driver.findElement(By.id("flow_one_page")).click();
        driver.findElement(By.id("flow_preserve_login_email")).clear();
        driver.findElement(By.id("flow_preserve_login_email")).sendKeys(username);
        driver.findElement(By.id("flow_preserve_login_password")).clear();
        driver.findElement(By.id("flow_preserve_login_password")).sendKeys(password);
        driver.findElement(By.id("flow_preserve_login_button")).click();
    }

    public static String getRandomString(int length) {

        String KeyString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
        StringBuffer sb = new StringBuffer();
        int len = KeyString.length();
        for (int i = 0; i < length; i++) {
            sb.append(KeyString.charAt((int) Math.round(Math.random() * (len - 1))));
        }
        return sb.toString();
    }
    @BeforeClass
    public void setUp() throws Exception {
//        System.setProperty("webdriver.chrome.driver", "F:/app/new/chromedriver.exe");
//        driver = new ChromeDriver();
//        baseUrl = "http://10.141.211.178:32412";
        driver = new RemoteWebDriver(new URL("http://hub:4444/wd/hub"),
                DesiredCapabilities.chrome());
        baseUrl = "http://10.141.211.162:30118";
        trainType = "0";//all
        driver.manage().timeouts().implicitlyWait(30, TimeUnit.SECONDS);
    }

    @Test
    //Test Flow Preserve Step 1: - Login
    public void testLogin1()throws Exception{
        driver.get(baseUrl + "/");

        //define username and password
        String username = "fdse_microservices@163.com";
        String password = "DefaultPassword";

        //call function login
        login(driver,username,password);
        Thread.sleep(50000);
        if( !getLoginStatus()){
            Thread.sleep(20000);
            System.out.println("#### Oh! Give me one more time! ##########");
            login(driver,username,password);
            Thread.sleep(20000);
            if( !getLoginStatus()){
                Thread.sleep(20000);
                System.out.println("#### Oh! Give me the second time! ##########");
                login(driver,username,password);
                Thread.sleep(20000);
                if( !getLoginStatus()){
                    throw new Exception("!!!!!!!!! Cannot login !!!!!!!!!");
                }
            }
        }


        //get login status
//        getLoginStatus();

        loginNumber = Integer.parseInt(driver.findElement(By.id("login-people-number")).getAttribute("value"));
        System.out.println("********* first login number: " + loginNumber);
        ((JavascriptExecutor)driver).executeScript("document.getElementById('flow_preserve_login_msg').innerText=''");

        ((JavascriptExecutor) driver).executeScript("arguments[0].scrollIntoView();",  driver.findElement(By.id("flow_preserve_login_button")));
        driver.findElement(By.id("flow_preserve_login_button")).click();
        Thread.sleep(10000);
        if( !getLoginStatus()){
            throw new Exception("!!!!!!!!! Cannot login !!!!!!!!!");
        }
        int a = Integer.parseInt(driver.findElement(By.id("login-people-number")).getAttribute("value"));
        System.out.println("********* second login number: " + a);
        ((JavascriptExecutor)driver).executeScript("document.getElementById('flow_preserve_login_msg').innerText=''");

        ((JavascriptExecutor) driver).executeScript("arguments[0].scrollIntoView();",  driver.findElement(By.id("flow_preserve_login_button")));
        driver.findElement(By.id("flow_preserve_login_button")).click();
        Thread.sleep(10000);
        if( !getLoginStatus()){
            throw new Exception("!!!!!!!!! Cannot login !!!!!!!!!");
        }
        int b = Integer.parseInt(driver.findElement(By.id("login-people-number")).getAttribute("value"));
        System.out.println("********* third login number: " + b);
        driver.findElement(By.id("login-people-number")).sendKeys("");
        ((JavascriptExecutor)driver).executeScript("document.getElementById('flow_preserve_login_msg').innerText=''");

        ((JavascriptExecutor) driver).executeScript("arguments[0].scrollIntoView();",  driver.findElement(By.id("flow_preserve_login_button")));
        driver.findElement(By.id("flow_preserve_login_button")).click();
        Thread.sleep(10000);
        if( !getLoginStatus()){
            throw new Exception("!!!!!!!!! Cannot login !!!!!!!!!");
        }

        int n = Integer.parseInt(driver.findElement(By.id("login-people-number")).getAttribute("value"));
        System.out.println("********* final login number: " + n);
        Assert.assertEquals( n, loginNumber+3);

    }


    private boolean getLoginStatus(){
        ((JavascriptExecutor) driver).executeScript("arguments[0].scrollIntoView();",  driver.findElement(By.id("flow_preserve_login_msg")));
        String statusLogin = driver.findElement(By.id("flow_preserve_login_msg")).getText();
        if("".equals(statusLogin))
            System.out.println("Failed to Login! Status is Null!");
        else if(statusLogin.startsWith("Success"))
            System.out.println("Success to Login! Status:"+statusLogin);
        else
            System.out.println("Failed to Login! Status:"+statusLogin);

        if(statusLogin.startsWith("Success")){
            return true;
        } else {
            return false;
        }
    }


    @AfterClass
    public void tearDown() throws Exception {
        driver.quit();
    }

}
