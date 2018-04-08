package test;

import backend.domain.TestFlowListener;
import backend.domain.TestFlowReporter;
import org.testng.Assert;
import org.testng.annotations.Listeners;
import org.testng.annotations.Test;

//@Listeners({TestFlowReporter.class})
//@Listeners({TestFlowListener.class})
public class SimpleTest {

    @Test
    public void testMethodOne() throws Exception {
        Assert.assertTrue(true);
        throw new Exception("233333");
    }

//    @Test
//    public void testMethodTwo() {
//        Assert.assertTrue(false);
//    }
//
//    @Test(dependsOnMethods={"testMethodTwo"})
//    public void testMethodThree() {
//
//        Assert.assertTrue(true);
//    }
}
