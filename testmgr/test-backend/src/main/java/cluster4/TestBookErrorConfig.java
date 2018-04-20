package cluster4;

import org.openqa.selenium.By;
import org.openqa.selenium.JavascriptExecutor;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
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

public class TestBookErrorConfig {
    private WebDriver driver;
    private String trainType;//0--all,1--GaoTie,2--others
    private String baseUrl;
    public static void login(WebDriver driver,String username,String password){
        ((JavascriptExecutor) driver).executeScript("arguments[0].scrollIntoView();",  driver.findElement(By.id("flow_one_page")));
        driver.findElement(By.id("flow_one_page")).click();
        driver.findElement(By.id("flow_preserve_login_email")).clear();
        driver.findElement(By.id("flow_preserve_login_email")).sendKeys(username);
        driver.findElement(By.id("flow_preserve_login_password")).clear();
        driver.findElement(By.id("flow_preserve_login_password")).sendKeys(password);
        driver.findElement(By.id("flow_preserve_login_button")).click();
    }
    //获取指定位数的随机字符串(包含数字,0<length)
    public static String getRandomString(int length) {
        //随机字符串的随机字符库
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
//        baseUrl = "http://10.141.211.180:30776";
        driver = new RemoteWebDriver(new URL("http://hub:4444/wd/hub"),
                DesiredCapabilities.chrome());
        baseUrl = "http://10.141.211.174:30082";
        trainType = "0";//all
        driver.manage().timeouts().implicitlyWait(10, TimeUnit.SECONDS);
    }

    @Test
    //Test Flow Preserve Step 1: - Login
    public void testLogin()throws Exception{
        driver.get(baseUrl + "/");

        //define username and password
        String username = "fdse_microservices@163.com";
        String password = "DefaultPassword";

        //call function login
        Thread.sleep(30000);
        login(driver,username,password);
        Thread.sleep(10000);
        if( !getLoginStatus()){
            Thread.sleep(5000);
            System.out.println("#### Oh! Give me one more time! ##########");
            login(driver,username,password);
            Thread.sleep(20000);
            if( !getLoginStatus()){
                Thread.sleep(5000);
                System.out.println("#### Oh! Give me the second time! ##########");
                login(driver,username,password);
                Thread.sleep(20000);
            }
        }

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

    @Test (dependsOnMethods = {"testLogin"})
    //test Flow Preserve Step 2: - Ticket Booking
    public void testBooking() throws Exception{
        //locate booking startingPlace input
        WebElement elementBookingStartingPlace = driver.findElement(By.id("travel_booking_startingPlace"));
        elementBookingStartingPlace.clear();
        elementBookingStartingPlace.sendKeys("Shang Hai");

        //locate booking terminalPlace input
        WebElement elementBookingTerminalPlace = driver.findElement(By.id("travel_booking_terminalPlace"));
        elementBookingTerminalPlace.clear();
        elementBookingTerminalPlace.sendKeys("Su Zhou");

        //locate booking Date input
        String bookDate = "";
        SimpleDateFormat sdf=new SimpleDateFormat("yyyy-MM-dd");
        Calendar newDate = Calendar.getInstance();
        Random randDate = new Random();
        int randomDate = randDate.nextInt(26);
        newDate.add(Calendar.DATE, randomDate+5);
        bookDate=sdf.format(newDate.getTime());

        JavascriptExecutor js = (JavascriptExecutor) driver;
        js.executeScript("document.getElementById('travel_booking_date').value='"+bookDate+"'");


        //locate Train Type input
        WebElement elementBookingTraintype = driver.findElement(By.id("search_select_train_type"));
        Select selTraintype = new Select(elementBookingTraintype);
        selTraintype.selectByValue(trainType); //ALL

        //locate Train search button
        WebElement elementBookingSearchBtn = driver.findElement(By.id("travel_booking_button"));
        elementBookingSearchBtn.click();
        Thread.sleep(30000);

        List<WebElement> ticketsList = driver.findElements(By.xpath("//table[@id='tickets_booking_list_table']/tbody/tr"));
        //Confirm ticket selection
        if (ticketsList.size() == 0) {
            elementBookingSearchBtn.click();
            ticketsList = driver.findElements(By.xpath("//table[@id='tickets_booking_list_table']/tbody/tr"));
        }
        if(ticketsList.size() > 0) {
            //Pick up a train at random and book tickets
            System.out.printf("Success to search tickets，the tickets list size is:%d%n",ticketsList.size());
//            Random rand = new Random();
//            int i = rand.nextInt(2000) % ticketsList.size();
//            WebElement elementBookingSeat = ticketsList.get(i).findElement(By.xpath("td[10]/select"));
//            Select selSeat = new Select(elementBookingSeat);
//            selSeat.selectByValue("3"); //2st
//            ticketsList.get(i).findElement(By.xpath("td[13]/button")).click();
//            Thread.sleep(10000);
        }
        else
            System.out.println("Tickets search failed!!!");
        Assert.assertEquals(ticketsList.size() > 0,true);
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
