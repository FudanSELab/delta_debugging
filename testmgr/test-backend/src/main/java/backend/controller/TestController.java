package backend.controller;

import backend.domain.*;
import backend.domain.delta.DeltaTestReporter;
import backend.domain.delta.DeltaTestRequest;
import backend.domain.delta.DeltaTestResponse;
import backend.domain.delta.DeltaTestResult;
import backend.domain.socket.SocketRequest;
import backend.domain.socket.SocketResponse;
import backend.domain.socket.SocketSessionRegistry;
import backend.service.ConfigService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.bind.annotation.*;
import org.testng.ITestNGListener;
import org.testng.TestNG;


import java.io.*;
import java.net.URISyntaxException;
import java.util.*;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.FutureTask;

@RestController
public class TestController {

    @Autowired
    ConfigService configService;

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/welcome", method = RequestMethod.GET)
    public String welcome() {
        return "hello";
    }

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/testBackend/reloadConfig", method = RequestMethod.GET)
    public String reload() {
        configService.reloadJson();
        return configService.getFileString();
    }

    /**session操作类*/
    @Autowired
    SocketSessionRegistry webAgentSessionRegistry;
    /**消息发送工具*/
    @Autowired
    private SimpMessagingTemplate template;

    @MessageMapping("/msg/testsingle")
    public void testsingle(SocketRequest message) throws Exception {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()) {
            String sessionId = webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            SocketResponse sr = new SocketResponse();
            if (!configService.containTestCase(message.getTestName())) {
                sr.setStatus(false);
                sr.setMessage("Test not exist");
                template.convertAndSendToUser(sessionId, "/topic/testresponse", sr, createHeaders(sessionId));
            } else {
                TestResponse result = runTest(message.getTestName());
                sr.setStatus(true);
                sr.setMessage("Test executed");
                sr.setTestResult(result);
                template.convertAndSendToUser(sessionId, "/topic/testresponse", sr, createHeaders(sessionId));
            }
        }
    }

    private MessageHeaders createHeaders(String sessionId) {
        SimpMessageHeaderAccessor headerAccessor = SimpMessageHeaderAccessor.create(SimpMessageType.MESSAGE);
        headerAccessor.setSessionId(sessionId);
        headerAccessor.setLeaveMutable(true);
        return headerAccessor.getMessageHeaders();
    }

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/testBackend/deltaTest", method = RequestMethod.POST)
    public DeltaTestResponse deltaTest(@RequestBody DeltaTestRequest request) throws Exception {
        List<String> testStrings = request.getTestNames();
        if(testStrings == null){
            DeltaTestResponse response = new DeltaTestResponse();
            response.setStatus(0);
            response.setMessage("The testString is null.");
            return response;
        }
        DeltaTestResponse response = new DeltaTestResponse();
        List<FutureTask<DeltaTestResult>> futureTasks = new ArrayList<FutureTask<DeltaTestResult>>();
        ExecutorService executorService = Executors.newFixedThreadPool(10);
        for(String s: testStrings){
            if( ! configService.containTestCase(s) ){
                response.setStatus(0);
                response.setMessage(s + " is not in the test list.");
                return response;
            } else {
                FutureTask<DeltaTestResult> futureTask = new FutureTask<DeltaTestResult>(new SingleDeltaTest(s));
                System.out.println("##############################################");
                System.out.println("############# add a new test task ############");
                System.out.println("##############################################");
                futureTasks.add(futureTask);
                executorService.submit(futureTask);
            }
        }
        int status = 1;
        String message = "Test all the chosen testcases";
        for (FutureTask<DeltaTestResult> futureTask : futureTasks) {
            response.addDeltaResult(futureTask.get());
            if( "EXCEPTION".equals(futureTask.get().getStatus())){
                status = -1;
                message = futureTask.get().getMessage();
                break;
            } else if( ! "SUCCESS".equals(futureTask.get().getStatus())){
                status = 0;
            }
        }
        response.setStatus(status);
        response.setMessage(message);
        // 清理线程池
        executorService.shutdown();
        return response;
    }

    //callable test
    class SingleDeltaTest implements Callable<DeltaTestResult>{

        private String testName;

        public SingleDeltaTest(String s){
            this.testName = s;
        }

        @Override
        public DeltaTestResult call() throws Exception {
            return runDeltaTest(this.testName);
        }
    }


    private DeltaTestResult runDeltaTest(String testString) throws Exception{
        TestNG testng = new TestNG();
        try{
            //must add the package name
            testng.setTestClasses(new Class[]{Class.forName("test." + testString)});
//        testng.setTestClasses(new Class[]{Class.forName(configService.getTestCase(testString))});
            DeltaTestReporter tfr = new DeltaTestReporter();
            testng.addListener((ITestNGListener)tfr);
            testng.setOutputDirectory("./test-output");
            testng.run();
            return tfr.getDeltaResult();
        } catch(Exception e){
            DeltaTestResult r = new DeltaTestResult();
            r.setStatus("EXCEPTION");
            r.setMessage("Cannot find and run test case: " + testString);
            return r;
        }
    }


    private TestResponse runTest(String testString) throws Exception{
        TestResponse response = new TestResponse();
        TestNG testng = new TestNG();
//        MyClassLoader mcl = new MyClassLoader(configService.getClassDir());
//        testng.setTestClasses(new Class[]{ mcl.loadClass(configService.getTestCase(testString)) });

        //must add the package name
        testng.setTestClasses(new Class[]{Class.forName("test." + testString)});
//        testng.setTestClasses(new Class[]{Class.forName(configService.getTestCase(testString))});

        TestFlowReporter tfr = new TestFlowReporter();
        testng.addListener((ITestNGListener)tfr);
        testng.setOutputDirectory("./test-output");
        testng.run();

        response.setResultList(tfr.getResultList());
        Integer[] count = tfr.getResultCount();
        response.setResultCount(count);
        if(count[1] ==0 && count[2] ==0 && count[3] ==0){
            response.setStatus(true);
            response.setMessage("All Passed");
        } else {
            response.setStatus(false);
            response.setMessage("Failed");
        }
        return response;
    }

    //get the directory structure of test cases
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/testBackend/getFileTree", method = RequestMethod.GET)
    public Map<String, List<String>> getFileTree() throws IOException, URISyntaxException, ClassNotFoundException {
//        ArrayList<FileDirectory> result = new ArrayList<FileDirectory>();
//        FileDirectory fd = new FileDirectory();
//        fd.setTitle("test");
//        fd.setType("folder");
//        System.out.println("filePath-------" + configService.getTestDir());
//        traverseFolder(configService.getTestDir(), fd);
//        result.add(fd);


//        ArrayList<FileDirectory> result = new ArrayList<FileDirectory>();
//        FileDirectory fd = new FileDirectory();
//        fd.setTitle("test");
//        fd.setType("folder");
//        for(String s: testCases){
//            FileNode fn = new FileNode();
//            fn.setTitle(s);
//            fn.setType("item");
//            fd.addProduct(fn);
//        }
//        result.add(fd);

        Map<String, List<String>>  testCases = configService.getTestFileNames();

        return testCases;
    }

    //    @Autowired
//    private SimpMessagingTemplate messagingTemplate;
//
//
//    @MessageMapping("/user/{userId}/send")
//    @SendToUser(value = "/send", broadcast = false)
//    public String send(@DestinationVariable String userId, SocketMessage message) throws Exception {
//        DateFormat df = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
//        message.date = df.format(new Date());
//        System.out.println("23333");
//        messagingTemplate.convertAndSendToUser(userId,"/send", message);
//        return "send";
//    }
//
//
//    @Scheduled(fixedRate = 1000)
//    @SendTo(value = "/getTime")
//    public Object getTime() throws Exception {
//        // 发现消息
//        DateFormat df = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
//        messagingTemplate.convertAndSend("/getTime", df.format(new Date()));
//        return "callback";
//    }

    //    /**
//     * right test
//     * @param request
//     * @return
//     * @throws Exception
//     */
//    @CrossOrigin(origins = "*")
//    @RequestMapping(value="/testBackend/test", method = RequestMethod.POST)
//    public  TestResponse test(@RequestBody TestRequest request) throws Exception {
//        String testString = request.getTestString();
//
//        if( ! configService.containTestCase(testString) ){
//            TestResponse response = new TestResponse();
//            response.setStatus(false);
//            response.setMessage("The file is not in the test list.");
//            return response;
//        }
//
//       return runTest(testString);
//    }

//    public void traverseFolder(String path, FileDirectory fd) throws IOException, URISyntaxException, ClassNotFoundException {
//        File file = new File(path);
//        System.out.println("filePath-------" + file.getAbsolutePath());
//
//        if (file.exists()) {
//            File[] files = file.listFiles();
//            if (files.length == 0) {
//                System.out.println("The file directory is empty!");
//                return;
//            } else {
//                for (File file2 : files) {
//                    if (file2.isDirectory()) {
//                        FileDirectory f2 = new FileDirectory();
//                        f2.setType("folder");
//                        f2.setTitle(file2.getName());
//                        fd.addProduct(f2);
//                        traverseFolder(file2.getAbsolutePath(), f2);
////                        System.out.println("File:" + file2.getAbsolutePath());
//                    } else {
//                        FileNode fn = new FileNode();
//                        fn.setType("item");
//                        fn.setTitle(file2.getName());
//                        fd.addProduct(fn);
////                        System.out.println("File Directory:" + file2.getAbsolutePath());
//                    }
//                }
//            }
//        } else {
//            System.out.println("The file is not exist!");
//        }
//    }

    //choose the testng listener
//    @CrossOrigin(origins = "*")
//    @RequestMapping(value="/testListener", method = RequestMethod.POST)
//    public  TestResponse testListener(@RequestBody TestRequest request) throws Exception {
//        String testString = request.getTestString();
//        TestResponse response = new TestResponse();
//
//        if( ! configService.containTestCase(testString) ){
//            response.setStatus(false);
//            response.setMessage("The file is not in the test list.");
//            return response;
//        }
//        TestNG testng = new TestNG();
//        MyClassLoader mcl = new MyClassLoader(configService.getClassDir());
//        testng.setTestClasses(new Class[]{ mcl.loadClass(configService.getTestCase(testString)) });
//        TestFlowListener tfl = new TestFlowListener();
//        testng.addListener((ITestNGListener)tfl);
//        testng.run();
//
//        response.setResultList(tfl.getResultList());
//        response.setStatus(true);
//        response.setMessage("Succsee");
//        return  response;
//    }

    //load the class across the network
    //    @CrossOrigin(origins = "*")
//    @RequestMapping(value="/loader", method = RequestMethod.GET)
//    public TestResponse loader() {
//        String rootUrl = "http://localhost:8080/";
//        NetworkClassLoader networkClassLoader = new NetworkClassLoader(rootUrl);
//        String classname = "test.SimpleTest";
//        Class clazz = null;
//        TestResponse response = new TestResponse();
//        try {
//            clazz = networkClassLoader.loadClass(classname);
//            System.out.println(clazz.getClassLoader());  //打印类加载器
////            Object newInstance = clazz.newInstance();
////            clazz.getMethod("getStr").invoke(newInstance);  //调用方法
//            TestNG testng = new TestNG();
//            testng.setTestClasses(new Class[]{clazz});
//            TestFlowReporter tfr = new TestFlowReporter();
//            testng.addListener((ITestNGListener)tfr);
//            testng.run();
//
//            response.setStatus(true);
//            response.setResultList(tfr.getResultList());
//            Integer[] count = tfr.getResultCount();
//            response.setResultCount(count);
//            if(count[1] ==0 && count[2] ==0 && count[3] ==0){
//                response.setMessage("All Passed");
//            } else {
//                response.setMessage("Failed");
//            }
//            return  response;
//        } catch (Exception e) {
//            e.printStackTrace();
//        }
//        return response;
//    }

}
