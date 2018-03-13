package test;

import org.openqa.selenium.By;
import org.openqa.selenium.JavascriptExecutor;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeDriver;
import org.openqa.selenium.remote.DesiredCapabilities;
import org.openqa.selenium.remote.RemoteWebDriver;
import org.springframework.core.annotation.Order;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

import java.net.URL;
import java.util.concurrent.TimeUnit;


public class TestServiceCancel {
    private WebDriver driver;
    private String baseUrl;
    public static void login(WebDriver driver,String username,String password){
        driver.findElement(By.id("flow_one_page")).click();
        driver.findElement(By.id("flow_preserve_login_email")).clear();
        driver.findElement(By.id("flow_preserve_login_email")).sendKeys(username);
        driver.findElement(By.id("flow_preserve_login_password")).clear();
        driver.findElement(By.id("flow_preserve_login_password")).sendKeys(password);
        driver.findElement(By.id("flow_preserve_login_button")).click();
    }
    @BeforeClass
    public void setUp() throws Exception {
//        System.setProperty("webdriver.chrome.driver", "/Users/hechuan/Downloads/chromedriver");
//        driver = new ChromeDriver();
        driver = new RemoteWebDriver(new URL("http://hub:4444/wd/hub"),
                DesiredCapabilities.chrome());
        baseUrl = "http://10.141.211.181:30004";
        driver.manage().timeouts().implicitlyWait(30, TimeUnit.SECONDS);
    }

    @Test
    public void login()throws Exception{
        driver.get(baseUrl + "/");

        //define username and password
        String username = "fdse_microservices@163.com";
        String password = "DefaultPassword";

        //call function login
        login(driver,username,password);
        Thread.sleep(1000);

        //get login status
        String statusLogin = driver.findElement(By.id("flow_preserve_login_msg")).getText();
        if("".equals(statusLogin))
            System.out.println("Failed to Login! Status is Null!");
        else if(statusLogin.startsWith("Success"))
            System.out.println("Success to Login! Status:"+statusLogin);
        else
            System.out.println("Failed to Login! Status:"+statusLogin);

        Assert.assertEquals(statusLogin.startsWith("Success"),true);
        driver.findElement(By.id("microservice_page")).click();
    }

    @Test (dependsOnMethods = {"login"})
    public void testCheckRefund() throws Exception{
        JavascriptExecutor js = (JavascriptExecutor) driver;
        js.executeScript("document.getElementById('single_cancel_order_id').value='5ad7750b-a68b-49c0-a8c0-32776b067703'");

        driver.findElement(By.id("single_cancel_refund_button")).click();
        Thread.sleep(500);
        String statusCancelRefundBtn = driver.findElement(By.id("single_cancel_refund_result")).getText();
        System.out.println("Cancel Refund Btn status:"+statusCancelRefundBtn);
        boolean flag = !"error".equals(statusCancelRefundBtn);

        if (flag)
            Assert.assertEquals(flag, true);
        else
            Assert.assertEquals(flag, false);
    }

    @Test (dependsOnMethods = {"testCheckRefund"})
    public void testTicketCancel() throws Exception {
        driver.findElement(By.id("single_cancel_button")).click();
        Thread.sleep(1000);
        String statusCancelOrderResult = driver.findElement(By.id("single_cancel_order_result")).getText();
        System.out.println("Do Cancel Btn status:"+statusCancelOrderResult);
        boolean flag = statusCancelOrderResult.startsWith("Success");

        if (flag)
            Assert.assertEquals(flag, true);
        else
            Assert.assertEquals(flag, false); // Order Status Cancel Not Permitted Check Refund: error
    }

    @AfterClass
    public void tearDown() throws Exception {
        driver.quit();
    }
}