package backend.domain.delta;

import org.testng.*;
import org.testng.xml.XmlSuite;

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Set;

public class DeltaTestReporter implements IReporter {

    List<ITestResult> l = new ArrayList<ITestResult>();
    DeltaTestResult deltaResult = new DeltaTestResult();

    public List<ITestResult> getResults(){
        return l;
    }

    @Override
    public  void generateReport(List<XmlSuite> list, List<ISuite> suites, String s) {
        l.clear();
        for (ISuite suite : suites) {
            Map<String, ISuiteResult> suiteResults = suite.getResults();
            for (ISuiteResult suiteResult : suiteResults.values()) {
                ITestContext testContext = suiteResult.getTestContext();
                IResultMap passedTests = testContext.getPassedTests();
                IResultMap failedTests = testContext.getFailedTests();
                IResultMap skippedTests = testContext.getSkippedTests();
                IResultMap failedConfig = testContext.getFailedConfigurations();
                l.addAll(this.listTestResult(passedTests));
                l.addAll(this.listTestResult(failedTests));
                l.addAll(this.listTestResult(skippedTests));
                l.addAll(this.listTestResult(failedConfig));
            }
        }
//        this.sort(l);
//        this.outputResult(l, s+"/test.txt");
        transferResult(l);
    }

    private void transferResult( List<ITestResult> list){
        if(list.size() > 0){
            deltaResult.setStatus("SUCCESS");
            deltaResult.setClassName(list.get(0).getTestClass().getName());
            for (ITestResult l : list) {
                if( ! "SUCCESS".equals(this.getStatus(l.getStatus())) ){
                    deltaResult.setStatus("FAILURE");
                    break;
                }
            }
            for (ITestResult l : list) {
                deltaResult.setDuration(deltaResult.getDuration() + l.getEndMillis() - l.getStartMillis() );
            }
        }
    }

    public DeltaTestResult getDeltaResult() {
        return deltaResult;
    }

    private ArrayList<ITestResult> listTestResult(IResultMap resultMap){
        Set<ITestResult> results = resultMap.getAllResults();
        return new ArrayList<ITestResult>(results);
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
