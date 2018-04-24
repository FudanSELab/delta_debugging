package deltabackend.domain.configDelta;

import java.util.List;

public class NewSingleDeltaCMResourceRequest {
    private String serviceName;
    private List<CMConfig> configs;

    public NewSingleDeltaCMResourceRequest(){

    }



    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public List<CMConfig> getConfigs() {
        return configs;
    }

    public void setConfigs(List<CMConfig> configs) {
        this.configs = configs;
    }
}
