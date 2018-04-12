package apiserver.util;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
@ConfigurationProperties(prefix="myConfig") //接收application.yml中的myConfig下面的属性
public class MyConfig {

    private List<Cluster> clusters;

    public MyConfig(){

    }

    public List<Cluster> getClusters() {
        return clusters;
    }

    public void setClusters(List<Cluster> clusters) {
        this.clusters = clusters;
    }
}
