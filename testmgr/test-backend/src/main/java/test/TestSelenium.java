package test;

import org.openqa.selenium.By;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeDriver;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

public class TestSelenium {

    private WebDriver driver;
    private String baseUrl;


    @BeforeClass
    public void setUp() throws Exception {
        System.setProperty("webdriver.chrome.driver", this.getClass().getResource("/").getPath() + "chromedriver.exe");
        driver = new ChromeDriver();
//        baseUrl = "http://10.141.212.24/";
        baseUrl = "http://10.141.211.164:30004";
//        driver.manage().timeouts().implicitlyWait(3, TimeUnit.SECONDS);
    }

    public static void login(WebDriver driver,String username,String password){
        driver.findElement(By.id("flow_one_page")).click();
        driver.findElement(By.id("flow_preserve_login_email")).clear();
        driver.findElement(By.id("flow_preserve_login_email")).sendKeys(username);
        driver.findElement(By.id("flow_preserve_login_password")).clear();
        driver.findElement(By.id("flow_preserve_login_password")).sendKeys(password);
        driver.findElement(By.id("flow_preserve_login_button")).click();
    }

    @Test
    public void testLogin()throws Exception{
        driver.get(baseUrl + "/");

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
    }

    @AfterClass
    public void tearDown() throws Exception {
        driver.quit();
    }
}
