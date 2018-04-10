package deltabackend.domain.configDelta;

import java.util.List;

public class ConfigDeltaRequest {

    String id;
    List<SingleDeltaCMResourceRequest> configs;
    List<String> tests;


    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public List<SingleDeltaCMResourceRequest> getConfigs() {
        return configs;
    }

    public void setConfigs(List<SingleDeltaCMResourceRequest> configs) {
        this.configs = configs;
    }

    public List<String> getTests() {
        return tests;
    }

    public void setTests(List<String> tests) {
        this.tests = tests;
    }


}
