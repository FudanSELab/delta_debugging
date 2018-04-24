package cluster4;

import helper.LoginInfo;
import helper.LoginResult;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestTemplate;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

import java.util.ArrayList;
import java.util.List;

import static org.springframework.http.HttpMethod.POST;

public class TestLoginErrorInstance2 {


    @BeforeClass
    public void setUp() throws Exception {
        //do nothing
        Thread.sleep(2000);
    }

    @Test
    public void testLogin()throws Exception{

        RestTemplate restTemplate = new RestTemplate();

        LoginInfo li = new LoginInfo();
        li.setPassword("DefaultPassword");
        li.setEmail("fdse_microservices@163.com");

        HttpHeaders headers = new HttpHeaders();
        List<String> cookies = new ArrayList<>();
        cookies.add("YsbCaptcha=233");
        headers.put(HttpHeaders.COOKIE,cookies);

        HttpEntity requestEntity = new HttpEntity(li, headers);
        //注意把这里换成你的集群的ip
        ResponseEntity<LoginResult> r = restTemplate.exchange("http://10.141.211.174:30082/login",HttpMethod.POST, requestEntity, LoginResult.class);
        LoginResult result = r.getBody();
        //[Error Process Seq] - 顺序没控制好的话result.message返回这个 status为false
        //Success.Processes Seq. - 顺序控制好了返回这个 status为true
        //Something Wrong - 其他不知道什么意外乱七八糟的情况返回这个,status为false
        System.out.println("~~~~LoginResult~~~~~ " + result.getMessage() );
        Assert.assertEquals(result.getMessage().contains("Success"),true);
        int loginNum = result.getLoginNum();
        if(loginNum > 0){
            ResponseEntity<LoginResult> r2 = restTemplate.exchange("http://10.141.211.174:30082/login",HttpMethod.POST, requestEntity, LoginResult.class);
            ResponseEntity<LoginResult> r3 = restTemplate.exchange("http://10.141.211.174:30082/login",HttpMethod.POST, requestEntity, LoginResult.class);
            ResponseEntity<LoginResult> r4 = restTemplate.exchange("http://10.141.211.174:30082/login",HttpMethod.POST, requestEntity, LoginResult.class);
            result = r4.getBody();
            Assert.assertEquals(result.getLoginNum(), loginNum+3);
            System.out.println(result.getLoginNum() + " VS " + (loginNum+3));
        } else {
            throw new Exception("get login number failed!");
        }
    }


    @AfterClass
    public void tearDown() throws Exception {
    }


}
