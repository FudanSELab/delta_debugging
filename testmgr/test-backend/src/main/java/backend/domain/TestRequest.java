package backend.domain;

import javax.validation.Valid;
import javax.validation.constraints.NotNull;

public class TestRequest {
    @Valid
    @NotNull
    private String testString;

    public TestRequest(){

    }
    public String getTestString() {
        return testString;
    }

    public void setTestString(String testString) {
        this.testString = testString;
    }



}
