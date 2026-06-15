import java.sql.*;
import java.nio.file.*;

public class ImageDumper {
    public static void main(String[] args) {
        if (args.length < 2) {
            System.out.println("Usage: java -cp hsqldb.jar;. ImageDumper <db_path_or_dir> <base_out_dir>");
            return;
        }
        
        Path dbArg = Paths.get(args[0]);
        String dbPath = dbArg.toString();

        Path baseDir = Paths.get(args[1]);
        String dbUrl = "jdbc:hsqldb:file:" + dbPath +
                       ";shutdown=true" +
                       ";readonly=true" +
                       ";hsqldb.files_readonly=true";
        System.out.println("Database URL: " + dbUrl);

        // Map index 0->3 to their respective directory names
        String[] fullDirs = {"rw_large1", "rw_large2", "rw_large3", "rw_large4"};
        String[] thumbDirs = {"rw_images1", "rw_images2", "rw_images3", "rw_images4"};

        try {
            Class.forName("org.hsqldb.jdbcDriver");
            try (Connection conn = DriverManager.getConnection(dbUrl, "SA", "")) {
                
                System.out.println("Extracting Full Images into rw_large1...rw_large4");
                extract(conn, "PUBLIC.FULL_IMAGES", baseDir, fullDirs);
                
                System.out.println("\nExtracting Thumbnails into rw_images1...rw_images4");
                extract(conn, "PUBLIC.THUMB_IMAGES", baseDir, thumbDirs);
                
                System.out.println("\nSUCCESS: All extractions complete.");
            }
        } catch (Exception e) {
            System.err.println("Fatal Database Error: " + e.getMessage());
        }
    }

    private static void extract(Connection conn, String tableName, Path baseDir, String[] subDirs) throws Exception {
        // Pre-create the targeted subdirectories
        for (String subDir : subDirs) {
            Files.createDirectories(baseDir.resolve(subDir));
        }

        String query = "SELECT ROW_ID, IMAGE_INDEX, IMAGE_DATA FROM " + tableName;
        
        try (Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(query)) {
            
            int count = 0;
            int errCount = 0;
            int skipCount = 0;
            
            while (rs.next()) {
                String rowId = rs.getString("ROW_ID");
                
                try {
                    int index = rs.getInt("IMAGE_INDEX");
                    
                    // Out-of-bounds safeguard (0 to 3)
                    if (index < 0 || index > 3) {
                        System.err.println("\n[Warning] Book ID " + rowId + " has out-of-bounds index " + index + " in " + tableName + ". Skipping.");
                        skipCount++;
                        continue;
                    }

                    byte[] data = rs.getBytes("IMAGE_DATA");
                    if (data == null || data.length == 0) continue;

                    // Detect format using magic numbers
                    String ext = ".dat"; // Safe fallback for unknown binary blobs
                    
                    if (data.length > 3) {
                        // JPG: FF D8 FF
                        if (data[0] == (byte)0xFF && (data[1] & 0xFF) == 0xD8 && data[2] == (byte)0xFF) {
                            ext = ".jpg";
                        } 
                        // PNG: 89 50 4E
                        else if (data[0] == (byte)0x89 && data[1] == (byte)0x50 && data[2] == (byte)0x4E) {
                            ext = ".png";
                        } 
                        // GIF: 47 49 46 ("GIF")
                        else if (data[0] == (byte)0x47 && data[1] == (byte)0x49 && data[2] == (byte)0x46) {
                            ext = ".gif";
                        }
                    }

                    // Write to the matched subdirectory using ROWID as the filename
                    Path outPath = baseDir.resolve(subDirs[index]).resolve(rowId + ext);
                    Files.write(outPath, data);
                    count++;
                    
                    // Progress indicator
                    if (count % 100 == 0) {
                        System.out.print(".");
                    }
                    
                } catch (Exception rowEx) {
                    errCount++;
                    System.err.println("\n[Warning] Skipped broken image for Book ID: " + rowId);
                }
            }
            
            System.out.println("\nSaved " + count + " files." + 
                (skipCount > 0 ? " (" + skipCount + " skipped bounds)" : "") + 
                (errCount > 0 ? " (" + errCount + " write errors)" : ""));
        }
    }
}