package backend.controller;

import backend.domain.*;
import backend.service.ConfigService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.testng.ITestNGListener;
import org.testng.TestNG;

import java.io.File;
import java.util.ArrayList;

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
    @RequestMapping(value="/reloadConfig", method = RequestMethod.GET)
    public String reload() {
        configService.reloadJson();
        return configService.getFileString();
    }

    /**
     * right test
     * @param request
     * @return
     * @throws Exception
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/test", method = RequestMethod.POST)
    public  TestResponse test(@RequestBody TestRequest request) throws Exception {
        String testString = request.getTestString();
        TestResponse response = new TestResponse();

        if( ! configService.containTestCase(testString) ){
            response.setStatus(false);
            response.setMessage("The file is not in the test list.");
            return response;
        }

        TestNG testng = new TestNG();
        MyClassLoader mcl = new MyClassLoader(configService.getClassDir());
        testng.setTestClasses(new Class[]{ mcl.loadClass(configService.getTestCase(testString)) });
        //attach the reporter to generate the result
        TestFlowReporter tfr = new TestFlowReporter();
        testng.addListener((ITestNGListener)tfr);
        //set the test output directory
        testng.setOutputDirectory("./testmgr/test-backend/test-output");
        testng.run();

        response.setStatus(true);
        response.setResultList(tfr.getResultList());
        Integer[] count = tfr.getResultCount();
        response.setResultCount(count);
        if(count[1] ==0 && count[2] ==0 && count[3] ==0){
            response.setMessage("All Passed");
        } else {
            response.setMessage("Failed");
        }
        return  response;
    }

    //get the directory structure of test cases
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/getFileTree", method = RequestMethod.GET)
    public ArrayList<FileDirectory> getFileTree(){
        ArrayList<FileDirectory> result = new ArrayList<FileDirectory>();
        FileDirectory fd = new FileDirectory();
        fd.setTitle("test");
        fd.setType("folder");
        traverseFolder(configService.getTestDir(), fd);
        result.add(fd);
        return result;
    }

    public void traverseFolder(String path, FileDirectory fd) {
        File file = new File(path);
        if (file.exists()) {
            File[] files = file.listFiles();
            if (files.length == 0) {
                System.out.println("The file directory is empty!");
                return;
            } else {
                for (File file2 : files) {
                    if (file2.isDirectory()) {
                        FileDirectory f2 = new FileDirectory();
                        f2.setType("folder");
                        f2.setTitle(file2.getName());
                        fd.addProduct(f2);
                        traverseFolder(file2.getAbsolutePath(), f2);
//                        System.out.println("File:" + file2.getAbsolutePath());
                    } else {
                        FileNode fn = new FileNode();
                        fn.setType("item");
                        fn.setTitle(file2.getName());
                        fd.addProduct(fn);
//                        System.out.println("File Directory:" + file2.getAbsolutePath());
                    }
                }
            }
        } else {
            System.out.println("The file is not exist!");
        }
    }

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
