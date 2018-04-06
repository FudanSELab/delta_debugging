package test;

import backend.domain.delta.DeltaTestReporter;
import backend.domain.delta.DeltaTestResult;
import org.testng.ITestNGListener;
import org.testng.TestNG;

public class Test {

    public  static void main(String[] args) throws Exception {
        TestNG testng = new TestNG();
        testng.setTestClasses(new Class[]{Class.forName("test.SimpleTest")});
        DeltaTestReporter tfr = new DeltaTestReporter();
        testng.addListener((ITestNGListener)tfr);
        testng.setOutputDirectory("./test-output");
        testng.run();
        DeltaTestResult r = tfr.getDeltaResult();
        System.out.println(r.getMessage() + "     " + r.getClassName() + "       " +  r.getStatus());
    }
}
