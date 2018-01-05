package backend.domain;

import org.testng.*;
import org.testng.xml.XmlSuite;

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Set;


public class TestFlowReporter implements IReporter {

    List<ITestResult> l = new ArrayList<ITestResult>();
    List<MyTestResult> resultList = new ArrayList<MyTestResult>();
    Integer[] resultCount = new Integer[4];//four result status

    public List<ITestResult> getResults(){
        return l;
    }

    @Override
    public  void generateReport(List<XmlSuite> list, List<ISuite> suites, String s) {
        l.clear();
        resultList.clear();
        resultCount[0] = resultCount[1] = resultCount[2] = resultCount[3] = 0;
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
        resultList.clear();
        for (ITestResult l : list) {
            MyTestResult m = new MyTestResult();
            m.setClassName(l.getTestClass().getRealClass().getName());
            m.setMethodName(l.getMethod().getMethodName());
            m.setStartTime(this.formatDate(l.getStartMillis()));
            m.setDuration(l.getEndMillis() - l.getStartMillis());
            m.setStatus(this.getStatus(l.getStatus()));
            resultList.add(m);
        }
    }

    public List<MyTestResult> getResultList(){
        return resultList;
    }

    public Integer[] getResultCount(){
        return resultCount;
    }

    private ArrayList<ITestResult> listTestResult(IResultMap resultMap){
        Set<ITestResult> results = resultMap.getAllResults();
        return new ArrayList<ITestResult>(results);
    }

    private String getStatus(int status){
        String statusString = null;
        switch (status) {
            case 1:
                resultCount[0]++;
                statusString = "SUCCESS";
                break;
            case 2:
                resultCount[1]++;
                statusString = "FAILURE";
                break;
            case 3:
                resultCount[2]++;
                statusString = "SKIP";
                break;
            default:
                resultCount[3]++;
                break;
        }
        return statusString;
    }

    private String formatDate(long date){
        SimpleDateFormat formatter = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
        return formatter.format(date);
    }

    //    private void sort(List<ITestResult> list){
//        Collections.sort(list, new Comparator<ITestResult>() {
//            @Override
//            public int compare(ITestResult r1, ITestResult r2) {
//                if(r1.getStartMillis()>r2.getStartMillis()){
//                    return 1;
//                }else{
//                    return -1;
//                }
//            }
//        });
//    }
//
//    private void outputResult(List<ITestResult> list, String path){
//        try {
//            BufferedWriter output = new BufferedWriter(new FileWriter(new File(path)));
//            StringBuffer sb = new StringBuffer();
//            for (ITestResult result : list) {
//                if(sb.length()!=0){
//                    sb.append("\r\n");
//                }
//                sb.append(result.getTestClass().getRealClass().getName())
//                        .append(" ")
//                        .append(result.getMethod().getMethodName())
//                        .append(" ")
//                        .append(this.formatDate(result.getStartMillis()))
//                        .append(" ")
//                        .append(result.getEndMillis()-result.getStartMillis())
//                        .append("ms ")
//                        .append(this.getStatus(result.getStatus()));
//            }
//            output.write(sb.toString());
//            output.flush();
//            output.close();
//        } catch (IOException e) {
//            e.printStackTrace();
//        }
//    }

}
