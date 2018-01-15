package test;

import org.openqa.selenium.By;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
import org.openqa.selenium.chrome.ChromeDriver;
import org.openqa.selenium.remote.DesiredCapabilities;
import org.openqa.selenium.remote.RemoteWebDriver;
import org.openqa.selenium.support.ui.ExpectedCondition;
import org.openqa.selenium.support.ui.WebDriverWait;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.DataProvider;
import org.testng.annotations.Test;

import java.net.MalformedURLException;
import java.net.URL;
import java.util.HashSet;
import java.util.Iterator;
import java.util.Set;

public class TestRemoteDriver {

    private WebDriver driver;
    private String baseUrl;

    @BeforeClass
    public void setUp() throws Exception {
        driver = new RemoteWebDriver(new URL("http://hub:4444/wd/hub"),
                    DesiredCapabilities.chrome());
        baseUrl = "http://www.baidu.com";
//        driver.manage().timeouts().implicitlyWait(3, TimeUnit.SECONDS);
    }


    @org.testng.annotations.Test
    public void testLogin()throws Exception{
        driver.get(baseUrl + "/");

        WebElement element = driver.findElement(By.name("wd"));
        element.sendKeys("test");
        element.submit();
        System.out.println("Page title is: " + driver.getTitle());
        (new WebDriverWait(driver, 10)).until(new ExpectedCondition<Boolean>() {
            public Boolean apply(WebDriver d) {
                return d.getTitle().toLowerCase().startsWith("test");
            }
        });
        System.out.println("Page title is: " + driver.getTitle());

//        Assert.assertEquals(statusLogin.startsWith("Success"),true);
    }

    @AfterClass
    public void tearDown() throws Exception {
        driver.quit();
    }
}
