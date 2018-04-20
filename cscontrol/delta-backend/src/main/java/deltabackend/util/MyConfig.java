package deltabackend.util;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
@ConfigurationProperties(prefix="myConfig") //接收application.yml中的myConfig下面的属性
public class MyConfig {

    private Map<String, String> zipkinUrl = new HashMap<>();
    private List<String> clusters = new ArrayList<String>();


    public List<String> getClusters() {
        return clusters;
    }

    public void setClusters(List<String> clusters) {
        this.clusters = clusters;
    }

    public Map<String, String> getZipkinUrl() {
        return zipkinUrl;
    }

    public void setZipkinUrl(Map<String, String> zipkinUrl) {
        this.zipkinUrl = zipkinUrl;
    }

}
