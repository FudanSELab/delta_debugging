package apiserver.util;

//import java.io.BufferedWriter;
//import java.io.File;
//import java.io.FileWriter;
//import java.io.IOException;

public class FileOperation {

//    public static boolean createFile(String destFileName) {
//        File file = new File(destFileName);
//        if(file.exists()) {
//            System.out.println("[=====]创建单个文件" + destFileName + "失败，目标文件已存在！");
//            return false;
//        }
//        if (destFileName.endsWith(File.separator)) {
//            System.out.println("[=====]创建单个文件" + destFileName + "失败，目标文件不能为目录！");
//            return false;
//        }
//        //判断目标文件所在的目录是否存在
//        if(!file.getParentFile().exists()) {
//            //如果目标文件所在的目录不存在，则创建父目录
//            System.out.println("[=====]目标文件所在目录不存在，准备创建它！");
//            if(!file.getParentFile().mkdirs()) {
//                System.out.println("[=====]创建目标文件所在目录失败！");
//                return false;
//            }
//        }
//        //创建目标文件
//        try {
//            if (file.createNewFile()) {
//                System.out.println("[=====]创建单个文件" + destFileName + "成功！");
//                return true;
//            } else {
//                System.out.println("[=====]创建单个文件" + destFileName + "失败！");
//                return false;
//            }
//        } catch (IOException e) {
//            e.printStackTrace();
//            System.out.println("[=====]创建单个文件" + destFileName + "失败！" + e.getMessage());
//            return false;
//        }
//    }

//    public static boolean  clearAndWriteFile(String filename,String svcName){
//        String str = "test";
//        try{
//            File file= new File(filename);
//            file.createNewFile(); // 创建新文件
//            BufferedWriter out = new BufferedWriter(new FileWriter(file));
//            out.write(str); // \r\n即为换行
//            out.flush(); // 把缓存区内容压入文件
//            out.close(); // 最后记得关闭文件
//            return true;
//        }catch(Exception e){
//            e.printStackTrace();
//            return false;
//        }
//    }
}
