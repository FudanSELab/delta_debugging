package cluster1;

import helper.QueryInfo;
import helper.TripResponse;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.web.client.HttpServerErrorException;
import org.springframework.web.client.RestTemplate;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;

public class TestSearchTicketConfig {

    @BeforeClass
    public void setUp() {
    }
    @Test
    public void test() throws Exception {

        RestTemplate restTemplate = new RestTemplate();
        QueryInfo queryInfo = new QueryInfo();
        queryInfo.setStartingPlace("Shang Hai");
        queryInfo.setEndPlace("Su Zhou");
        String string = "2018-04-28";
        SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd");
        queryInfo.setDepartureTime(sdf.parse(string));

        HttpHeaders headers = new HttpHeaders();
        List<String> cookies = new ArrayList<String>();
        cookies.add("loginId=4d2a46c7-71cb-4cf1-b5bb-b68406d9da6f");
        headers.put(HttpHeaders.COOKIE,cookies);

        HttpEntity requestEntity = new HttpEntity(queryInfo, headers);

        try{
            Thread.sleep(10000);
            ResponseEntity<ArrayList<TripResponse>> r = restTemplate.exchange("http://10.141.211.179:30224/travel/query",HttpMethod.POST, requestEntity, new ParameterizedTypeReference<ArrayList<TripResponse>>(){});
            ArrayList<TripResponse> result = r.getBody();
            System.out.println(result);
            if(null == result || result.size() <= 0){
                throw new Exception("Something wrong");
            }
//            Assert.assertEquals((null != result) && (result.size() > 0), true);
        } catch (HttpServerErrorException e){
            if(e.getRawStatusCode() == 500){
                Assert.assertEquals(1,0);
            }
        }

    }

    @AfterClass
    public void tearDown() {

    }
}
