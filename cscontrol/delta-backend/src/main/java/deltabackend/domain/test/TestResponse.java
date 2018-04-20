package deltabackend.domain.test;

import deltabackend.domain.test.MyTestResult;

import java.util.List;

public class TestResponse {

    private boolean status;
    private String message;
    private Integer[] resultCount;
    private List<MyTestResult> resultList;

    public TestResponse(){

    }

    public Integer[] getResultCount() {
        return resultCount;
    }

    public void setResultCount(Integer[] resultCount) {
        this.resultCount = resultCount;
    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public List<MyTestResult> getResultList() {
        return resultList;
    }

    public void setResultList(List<MyTestResult> resultList) {
        this.resultList = resultList;
    }

}
