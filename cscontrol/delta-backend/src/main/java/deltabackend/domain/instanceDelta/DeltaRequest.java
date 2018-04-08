package deltabackend.domain.instanceDelta;

import java.util.List;

public class DeltaRequest {

    String id;
    List<String> env;
    List<String> tests;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }
    public List<String> getEnv() {
        return env;
    }

    public void setEnv(List<String> env) {
        this.env = env;
    }

    public List<String> getTests() {
        return tests;
    }

    public void setTests(List<String> tests) {
        this.tests = tests;
    }

}
