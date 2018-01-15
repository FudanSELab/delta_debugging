package test;

import org.openqa.selenium.By;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeDriver;
import org.openqa.selenium.remote.DesiredCapabilities;
import org.openqa.selenium.remote.RemoteWebDriver;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.DataProvider;
import org.testng.annotations.Test;

import java.net.URL;
import java.util.HashSet;
import java.util.Iterator;
import java.util.Set;

public class TestSelenium {

    private WebDriver driver;
    private String baseUrl;


    @BeforeClass
    public void setUp() throws Exception {
//        System.setProperty("webdriver.chrome.driver", this.getClass().getResource("/").getPath() + "chromedriver.exe");
//        System.setProperty("webdriver.chrome.driver",  "/app/chromedriver.exe");
//        driver = new ChromeDriver();
//        baseUrl = "http://10.141.212.24/";
        driver = new RemoteWebDriver(new URL("http://hub:4444/wd/hub"),
                DesiredCapabilities.chrome());
        baseUrl = "http://10.141.211.164:30004";
//        driver.manage().timeouts().implicitlyWait(3, TimeUnit.SECONDS);
    }

//    @DataProvider(name="dataSource")
//    //返回Iterator<Object[]>的数据驱动
//    public Iterator<Object[]> getdata() {
//        Set<Object[]> set = new HashSet<Object[]>();
//        set.add(new String[]{"fdse_microservices@163.com","DefaultPassword"});
//        set.add(new String[]{"fdse_microservices@163.com","Default"});
//        return set.iterator();
//    }

    public static void login(WebDriver driver,String username,String password){
        driver.findElement(By.id("flow_one_page")).click();
        driver.findElement(By.id("flow_preserve_login_email")).clear();
        driver.findElement(By.id("flow_preserve_login_email")).sendKeys(username);
        driver.findElement(By.id("flow_preserve_login_password")).clear();
        driver.findElement(By.id("flow_preserve_login_password")).sendKeys(password);
        driver.findElement(By.id("flow_preserve_login_button")).click();
    }


//    @Test(dataProvider="dataSource")
    @Test
    public void testLogin()throws Exception{
        driver.get(baseUrl + "/");

        String username = "fdse_microservices@163.com";
        String password = "DefaultPassword";
//        String username = "fdse_microservices@163.com";
//        String password = "WrongPassword";

        //call function login
        login(driver,username,password);
//        login(driver,a,b);
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
