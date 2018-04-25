package cluster6;

import helper.CancelOrderResult;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.web.client.RestTemplate;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

public class TestF12 {
    @BeforeClass
    public void setUp() throws Exception {
        //do nothing
    }

    @Test
    public void testLogin()throws Exception {

        RestTemplate restTemplate = new RestTemplate();

        //首先向OrderService发送请求，锁定车站
        HttpEntity requestEntity = new HttpEntity(null, new HttpHeaders());
        ResponseEntity<Boolean> re = restTemplate.exchange(
                "http://10.141.211.161:30112/adminOrder/suspendOrder/shanghai/nanjing",
                HttpMethod.GET,
                requestEntity,
                Boolean.class);

        //确保请求被执行完毕
        System.out.println("锁定车站结果：" + re.getBody().booleanValue());
        Assert.assertEquals(re.getBody().booleanValue(), true);

        //发出十个退票请求，每次间隔十秒
        for (int i = 0; i < 10; i++) {
            //停顿十秒，给这个辣鸡负载均衡一些反应时间
            Thread.sleep(10000);
            ResponseEntity<CancelOrderResult> cancel = restTemplate.exchange(
                    "http://10.141.211.161:30185/cancelOrder/5ad7750b-a68b-49c0-a8c0-32776b067703",
                    HttpMethod.GET,
                    requestEntity,
                    CancelOrderResult.class);
            System.out.println("退订车票：" + cancel.getBody());
            Assert.assertEquals(cancel.getBody() == null || cancel.getBody().getMessage().length() < 2, true);
        }

    }


    @AfterClass
    public void tearDown() throws Exception {
        //把锁定的车票解除掉
        RestTemplate restTemplate = new RestTemplate();
        HttpEntity requestEntity = new HttpEntity(null, new HttpHeaders());
        for (int i = 0; i < 10; i++) {
            Thread.sleep(5000);
            ResponseEntity<Boolean> cancel = restTemplate.exchange(
                    "http://10.141.211.161:30112/adminOrder/cancelSuspendOrder/shanghai/nanjing",
                    HttpMethod.GET,
                    requestEntity,
                    Boolean.class);
        }
    }
}
