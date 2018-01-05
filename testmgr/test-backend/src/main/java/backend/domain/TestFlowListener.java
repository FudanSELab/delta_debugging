package backend.domain;

import org.testng.ITestResult;
import org.testng.TestListenerAdapter;

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.List;

public class TestFlowListener extends TestListenerAdapter {

    private int m_count = 0;
    List<MyTestResult> resultList = new ArrayList<MyTestResult>();

    @Override
    public void onTestFailure(ITestResult tr) {
       addResult(tr);
        log(tr.getName()+ "--Test method failed\n");
    }

    @Override
    public void onTestSkipped(ITestResult tr) {
        addResult(tr);
        log(tr.getName()+ "--Test method skipped\n");
    }

    @Override
    public void onTestSuccess(ITestResult tr) {
        addResult(tr);
        log(tr.getName()+ "--Test method success\n");
    }

    private void log(String string) {
        System.out.print(string);
        if (++m_count % 40 == 0) {
            System.out.println("");
        }
    }

    private void addResult(ITestResult tr){
        MyTestResult m = new MyTestResult();
        m.setClassName(tr.getTestClass().getRealClass().getName());
        m.setMethodName(tr.getMethod().getMethodName());
        m.setStartTime(this.formatDate(tr.getStartMillis()));
        m.setDuration(tr.getEndMillis() - tr.getStartMillis());
        m.setStatus(this.getStatus(tr.getStatus()));
        resultList.add(m);
    }

    public List<MyTestResult> getResultList(){
        return resultList;
    }

    private String getStatus(int status){
        String statusString = null;
        switch (status) {
            case 1:
                statusString = "SUCCESS";
                break;
            case 2:
                statusString = "FAILURE";
                break;
            case 3:
                statusString = "SKIP";
                break;
            default:
                break;
        }
        return statusString;
    }

    private String formatDate(long date){
        SimpleDateFormat formatter = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
        return formatter.format(date);
    }
}
