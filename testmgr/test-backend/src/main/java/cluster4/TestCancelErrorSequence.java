package cluster4;

import org.openqa.selenium.*;
import org.openqa.selenium.remote.DesiredCapabilities;
import org.openqa.selenium.remote.RemoteWebDriver;
import org.openqa.selenium.support.ui.Select;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

import java.net.URL;
import java.text.SimpleDateFormat;
import java.util.Calendar;
import java.util.List;
import java.util.Random;
import java.util.concurrent.TimeUnit;

public class TestCancelErrorSequence {
    private WebDriver driver;
    private String baseUrl;
    private List<WebElement> myOrdersList;
    private String trainType;//0--all,1--GaoTie,2--others
    public static void login(WebDriver driver, String username, String password){
        ((JavascriptExecutor) driver).executeScript("arguments[0].scrollIntoView();",  driver.findElement(By.id("flow_one_page")));
        driver.findElement(By.id("flow_one_page")).click();
        driver.findElement(By.id("flow_preserve_login_email")).clear();
        driver.findElement(By.id("flow_preserve_login_email")).sendKeys(username);
        driver.findElement(By.id("flow_preserve_login_password")).clear();
        driver.findElement(By.id("flow_preserve_login_password")).sendKeys(password);
        driver.findElement(By.id("flow_preserve_login_button")).click();
    }
    @BeforeClass
    public void setUp() throws Exception {
//                System.setProperty("webdriver.chrome.driver", "F:/app/new/chromedriver.exe");
//        driver = new ChromeDriver();
//        baseUrl = "http://10.141.211.180:30776";
        driver = new RemoteWebDriver(new URL("http://hub:4444/wd/hub"),
                DesiredCapabilities.chrome());
        baseUrl = "http://10.141.211.174:30082";
        trainType = "2";
        driver.manage().timeouts().implicitlyWait(30, TimeUnit.SECONDS);
    }
    @Test
    //Test Flow Preserve Step 1: - Login
    public void testLogin()throws Exception{
        driver.get(baseUrl + "/");

        //define username and password
        String username = "fdse_microservices@163.com";
        String password = "DefaultPassword";

        //call function login
        login(driver,username,password);
        Thread.sleep(10000);
        if( !getLoginStatus()){
            Thread.sleep(5000);
            System.out.println("#### Oh! Give me one more time! ##########");
            login(driver,username,password);
            Thread.sleep(20000);
        }

        //get login status
        String statusLogin = driver.findElement(By.id("flow_preserve_login_msg")).getText();
        if("".equals(statusLogin)) {
            System.out.println("Failed to Login! Status is Null!");
        } else if(statusLogin.startsWith("Success")) {
            System.out.println("Success to Login! Status:" + statusLogin);
        } else {
            System.out.println("Failed to Login! Status:" + statusLogin);
        }
        Assert.assertEquals(statusLogin.startsWith("Success"),true);



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

    @Test (dependsOnMethods = {"testLogin"})
    public void testViewOrders() throws Exception{
//        Thread.sleep(1000);
//        String js = "document.getElementById('flow_two_page').scrollIntoView(false)";
//        ((JavascriptExecutor)driver).executeScript(js);
        Thread.sleep(1000);
        ((JavascriptExecutor) driver).executeScript("arguments[0].scrollIntoView();",  driver.findElement(By.id("flow_two_page")));
        driver.findElement(By.id("flow_two_page")).click();

        driver.findElement(By.id("refresh_my_order_list_button")).click();
        Thread.sleep(1000);
        //gain my oeders
        myOrdersList = driver.findElements(By.xpath("//div[@id='my_orders_result']/div"));
        if (myOrdersList.size() > 0) {
            System.out.printf("Success to show my orders list，the list size is:%d%n",myOrdersList.size());
        } else {
            System.out.println("Failed to show my orders list，the list size is 0 or No orders in this user!");
        }
        Assert.assertEquals(myOrdersList.size() > 0,true);
    }


    @Test (dependsOnMethods = {"testViewOrders"})
    public void testClickOrderCancel() throws Exception{
        System.out.printf("The orders list size is:%d%n",myOrdersList.size());
        String statusOrder  = "";
        int i;
        //Find the first not paid order .
        for(i = 0;i < myOrdersList.size();i++) {
            //while(!(statusOrder.startsWith("Not")) && i < myOrdersList.size()) {
            //statusOrder = myOrdersList.get(i).findElement(By.xpath("/div[2]/div/div/form/div[7]/div/label[2]")).getText();
            statusOrder = myOrdersList.get(i).findElement(By.xpath("div[2]//form[@role='form']/div[7]/div/label[2]")).getText();
            if(statusOrder.startsWith("Paid"))
                break;
        }
        if(i == myOrdersList.size() || i > myOrdersList.size())
            System.out.printf("Failed,there is no not paid order!");
        Assert.assertEquals(i < myOrdersList.size(),true);

        Thread.sleep(1000);
        String js = "document.getElementsByClassName('ticket_cancel_btn')[0].scrollIntoView(false)";
        ((JavascriptExecutor)driver).executeScript(js);
        Thread.sleep(1000);

        myOrdersList.get(i).findElement(
                By.xpath("div[2]//form[@role='form']/div[12]/div/button[@class='ticket_cancel_btn btn btn-primary']")
        ).click();
        Thread.sleep(2000);

        driver.findElement(By.id("ticket_cancel_panel_confirm")).click();

        Thread.sleep(15000);

        try{
            Alert javascriptConfirm = driver.switchTo().alert();
            String statusAlert = driver.switchTo().alert().getText();
            System.out.println("The Alert information of Cancel Ticket："+statusAlert);
            Assert.assertEquals(statusAlert.startsWith("Success"),true);
            javascriptConfirm.accept();
        } catch(NoAlertPresentException e){
            Assert.assertEquals(0,1);
        }

    }



    @AfterClass
    public void tearDown() throws Exception {
        driver.quit();
    }
}
