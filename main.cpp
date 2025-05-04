// ░█▀▀░█▀█░█▀▄░█▀▀░█▀▄░█▀█░█▀▀░█▀▀░░░█▀▀░█░█░█▀█░█▀█░█▀▄░▀█▀░█▀▀░█▀▄
// ░█░░░█░█░█░█░█▀▀░█▀▄░█▀█░▀▀█░█▀▀░░░█▀▀░▄▀▄░█▀▀░█░█░█▀▄░░█░░█▀▀░█▀▄
// ░▀▀▀░▀▀▀░▀▀░░▀▀▀░▀▀░░▀░▀░▀▀▀░▀▀▀░░░▀▀▀░▀░▀░▀░░░▀▀▀░▀░▀░░▀░░▀▀▀░▀░▀

// Định nghĩa các macro để sử dụng Unicode trong Windows API
#define _UNICODE

// Bao gồm các thư viện cần thiết
#include <windows.h>         // Thư viện Windows API cơ bản
#include "json.hpp"          // Thư viện để xử lý JSON
using json = nlohmann::json; // Alias for convenience
#include <shlobj.h>          // Thư viện cho các hộp thoại chọn thư mục
#include <commctrl.h>        // Thư viện cho các control phổ biến như status bar
#include <iostream>          // Thư viện nhập/xuất chuẩn
#include <fstream>           // Thư viện để đọc/ghi file
#include <filesystem>        // Thư viện làm việc với hệ thống file
#include <string>            // Thư viện chuỗi
#include <vector>            // Thư viện mảng động
#include <sstream>           // Thư viện xử lý chuỗi
#include <algorithm>         // Thư viện các thuật toán
#include <set>               // Thư viện tập hợp

// Chỉ thị cho trình biên dịch link với thư viện Common Controls
#pragma comment(lib, "Comctl32.lib")
// Chỉ thị cho trình biên dịch sử dụng phiên bản 6.0 của Common Controls
#pragma comment(linker, "\"/manifestdependency:type='win32' name='Microsoft.Windows.Common-Controls' version='6.0.0.0' processorArchitecture='*' publicKeyToken='6595b64144ccf1df' language='*'\"")

// Tạo namespace ngắn hơn cho std::filesystem
namespace fs = std::filesystem;

// Hàm chuyển đổi từ std::string (chuỗi ANSI/UTF-8) sang std::wstring (chuỗi Unicode)
std::wstring StringToWideString(const std::string &str)
{
    // Nếu chuỗi rỗng, trả về chuỗi rỗng
    if (str.empty())
        return std::wstring();

    // Xác định kích thước cần thiết cho chuỗi Unicode
    int size_needed = MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), NULL, 0);
    // Tạo chuỗi Unicode với kích thước đã xác định
    std::wstring wstr(size_needed, 0);
    // Chuyển đổi chuỗi từ UTF-8 sang Unicode
    MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), &wstr[0], size_needed);
    return wstr;
}

// Hàm chuyển đổi từ std::wstring (chuỗi Unicode) sang std::string (chuỗi ANSI/UTF-8)
std::string WideStringToString(const std::wstring &wstr)
{
    // Nếu chuỗi rỗng, trả về chuỗi rỗng
    if (wstr.empty())
        return std::string();

    // Xác định kích thước cần thiết cho chuỗi UTF-8
    int size_needed = WideCharToMultiByte(CP_UTF8, 0, &wstr[0], (int)wstr.size(), NULL, 0, NULL, NULL);
    // Tạo chuỗi UTF-8 với kích thước đã xác định
    std::string str(size_needed, 0);
    // Chuyển đổi chuỗi từ Unicode sang UTF-8
    WideCharToMultiByte(CP_UTF8, 0, &wstr[0], (int)wstr.size(), &str[0], size_needed, NULL, NULL);
    return str;
}

// Khai báo trước các hàm sẽ được định nghĩa sau
void build_tree_structure(const fs::path &path, std::ostream &out, std::string prefix = "");
void dump_file(const fs::path &filepath, std::ostream &out);
bool should_read_file(const fs::path &filepath, const std::set<std::string> &extensions);
void process_directory(const fs::path &dir_path, const std::string &output_file, const std::set<std::string> &extensions);
std::vector<std::string> split(const std::string &s, char delimiter);

// Biến toàn cục để lưu trữ handle của các control UI
HWND g_hwndMain = NULL;       // Handle của cửa sổ chính
HWND g_hwndDirPath = NULL;    // Handle của ô nhập đường dẫn thư mục dự án
HWND g_hwndFilePath = NULL;   // Handle của ô nhập đường dẫn thư mục xuất file
HWND g_hwndExtensions = NULL; // Handle của ô nhập các phần mở rộng file
HWND g_hwndStatus = NULL;     // Handle của thanh trạng thái
std::string g_projectDir;     // Đường dẫn thư mục dự án
std::string g_outputFile;     // Đường dẫn thư mục xuất file

// Hàm hiển thị hộp thoại chọn thư mục dự án
bool BrowseForFolder(HWND hwndOwner, std::string &folderPath)
{
    // Khởi tạo cấu trúc BROWSEINFO với các tham số cần thiết
    BROWSEINFO bi = {0};
    bi.hwndOwner = hwndOwner;                               // Cửa sổ cha
    bi.lpszTitle = L"Select Project Directory";             // Tiêu đề hộp thoại
    bi.ulFlags = BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE; // Các cờ

    // Hiển thị hộp thoại chọn thư mục và lấy kết quả
    LPITEMIDLIST pidl = SHBrowseForFolder(&bi);
    if (pidl != NULL)
    {
        // Nếu người dùng đã chọn thư mục, lấy đường dẫn
        TCHAR path[MAX_PATH];
        if (SHGetPathFromIDList(pidl, path))
        {
            // Chuyển đường dẫn từ Unicode sang UTF-8
            folderPath = WideStringToString(std::wstring(path));
            // Giải phóng bộ nhớ đã cấp phát
            CoTaskMemFree(pidl);
            return true;
        }
        CoTaskMemFree(pidl);
    }
    return false;
}

// Hàm xử lý đặc biệt cho file Jupyter Notebook (.ipynb)
void process_jupyter_notebook(const fs::path &filepath, std::ostream &out)
{
    std::ifstream file(filepath);
    if (!file.is_open())
    {
        out << "\n// File: " << filepath.string() << " (could not be opened)\n";
        return;
    }

    try
    {
        // Đọc nội dung JSON từ file
        json notebook;
        file >> notebook;
        file.close();

        // Ghi tên file vào file output
        out << "\n// File: " << filepath.string() << " (Jupyter Notebook)\n";

        // Kiểm tra nếu có metadata
        if (notebook.contains("metadata"))
        {
            out << "// Notebook Metadata:\n";
            if (notebook["metadata"].contains("kernelspec") &&
                notebook["metadata"]["kernelspec"].contains("display_name"))
            {
                out << "// Kernel: " << notebook["metadata"]["kernelspec"]["display_name"].get<std::string>() << "\n";
            }
        }

        // Xử lý từng cell trong notebook
        if (notebook.contains("cells") && notebook["cells"].is_array())
        {
            out << "\n// Notebook Content:\n";
            int cell_number = 1;

            for (const auto &cell : notebook["cells"])
            {
                // Lấy loại cell (code, markdown, etc.)
                std::string cell_type = cell.contains("cell_type") ? cell["cell_type"].get<std::string>() : "unknown";

                out << "\n// --- Cell " << cell_number << " (" << cell_type << ") ---\n";

                // Xử lý nội dung cell
                if (cell.contains("source") && cell["source"].is_array())
                {
                    // Nối tất cả các dòng trong source
                    for (const auto &line : cell["source"])
                    {
                        if (line.is_string())
                        {
                            out << line.get<std::string>();
                        }
                    }
                    out << "\n";
                }

                // Nếu là cell code và có output
                if (cell_type == "code" && cell.contains("outputs") && cell["outputs"].is_array())
                {
                    out << "// --- Output ---\n";
                    for (const auto &output : cell["outputs"])
                    {
                        // Xử lý các loại output khác nhau
                        if (output.contains("output_type"))
                        {
                            std::string output_type = output["output_type"].get<std::string>();

                            if (output_type == "execute_result" || output_type == "display_data")
                            {
                                if (output.contains("data") &&
                                    output["data"].contains("text/plain") &&
                                    output["data"]["text/plain"].is_array())
                                {
                                    out << "// Result (text/plain):\n";
                                    for (const auto &line : output["data"]["text/plain"])
                                    {
                                        if (line.is_string())
                                        {
                                            out << "// " << line.get<std::string>() << "\n";
                                        }
                                    }
                                }
                            }
                            else if (output_type == "stream" &&
                                     output.contains("text") &&
                                     output["text"].is_array())
                            {
                                std::string stream_name = output.contains("name") ? output["name"].get<std::string>() : "stdout";
                                out << "// Stream (" << stream_name << "):\n";
                                for (const auto &line : output["text"])
                                {
                                    if (line.is_string())
                                    {
                                        out << "// " << line.get<std::string>() << "\n";
                                    }
                                }
                            }
                        }
                    }
                }

                cell_number++;
            }
        }
    }
    catch (const json::exception &e)
    {
        out << "// Error parsing notebook: " << e.what() << "\n";
    }
    catch (const std::exception &e)
    {
        out << "// Error processing notebook: " << e.what() << "\n";
    }
}

// Hàm hiển thị hộp thoại chọn thư mục xuất file
bool BrowseForOutputDirectory(HWND hwndOwner, std::string &directoryPath)
{
    // Sử dụng lại cơ chế tương tự hàm BrowseForFolder
    BROWSEINFO bi = {0};
    bi.hwndOwner = hwndOwner;
    bi.lpszTitle = L"Select Output Directory";
    bi.ulFlags = BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE;

    LPITEMIDLIST pidl = SHBrowseForFolder(&bi);
    if (pidl != NULL)
    {
        TCHAR path[MAX_PATH];
        if (SHGetPathFromIDList(pidl, path))
        {
            directoryPath = WideStringToString(std::wstring(path));
            CoTaskMemFree(pidl);
            return true;
        }
        CoTaskMemFree(pidl);
    }
    return false;
}

// Hàm xử lý chính của chương trình, chạy trong một thread riêng biệt
DWORD WINAPI ProcessDirectoryThread(LPVOID lpParam)
{
    // Hiển thị thông báo đang xử lý trên thanh trạng thái
    SendMessage(g_hwndStatus, SB_SETTEXT, 0, (LPARAM)L"Processing... Please wait.");

    // Lấy danh sách các phần mở rộng file từ ô nhập liệu
    wchar_t extensionsBuffer[1024];
    GetWindowText(g_hwndExtensions, extensionsBuffer, sizeof(extensionsBuffer) / sizeof(wchar_t));
    std::string extensionsStr = WideStringToString(extensionsBuffer);

    // Chuyển đổi chuỗi các phần mở rộng thành tập hợp
    std::set<std::string> extensions;
    if (!extensionsStr.empty())
    {
        // Tách chuỗi phần mở rộng theo dấu phẩy
        std::vector<std::string> ext_list = split(extensionsStr, ',');
        for (auto &ext : ext_list)
        {
            // Loại bỏ các khoảng trắng thừa
            ext.erase(0, ext.find_first_not_of(" \t\n\r\f\v"));
            ext.erase(ext.find_last_not_of(" \t\n\r\f\v") + 1);

            // Loại bỏ dấu chấm ở đầu nếu có
            if (!ext.empty() && ext[0] == '.')
            {
                ext = ext.substr(1);
            }

            // Thêm phần mở rộng vào tập hợp
            if (!ext.empty())
                extensions.insert(ext);
        }
    }

    try
    {
        // Tạo đường dẫn đầy đủ cho file output.txt trong thư mục đã chọn
        fs::path outputPath = fs::path(g_outputFile) / "output.txt";
        std::string fullOutputPath = outputPath.string();

        // Xử lý thư mục và tạo file output
        process_directory(g_projectDir, fullOutputPath, extensions);

        // Hiển thị thông báo thành công trên thanh trạng thái
        std::wstring statusMsg = L"Content exported to: " + StringToWideString(fullOutputPath);
        SendMessage(g_hwndStatus, SB_SETTEXT, 0, (LPARAM)statusMsg.c_str());
    }
    catch (const std::exception &e)
    {
        // Hiển thị thông báo lỗi nếu có
        std::wstring errorMsg = L"Error: " + StringToWideString(e.what());
        SendMessage(g_hwndStatus, SB_SETTEXT, 0, (LPARAM)errorMsg.c_str());
    }

    return 0;
}

// Hàm xử lý các sự kiện (message) của cửa sổ
LRESULT CALLBACK WndProc(HWND hwnd, UINT msg, WPARAM wParam, LPARAM lParam)
{
    switch (msg)
    {
    case WM_CREATE: // Khi cửa sổ được tạo
    {
        // Khởi tạo thư viện Common Controls
        INITCOMMONCONTROLSEX icc;
        icc.dwSize = sizeof(icc);
        icc.dwICC = ICC_WIN95_CLASSES;
        InitCommonControlsEx(&icc);

        // Tạo các control trên giao diện

        // Label "Project Directory:"
        CreateWindow(L"STATIC", L"Project Directory:", WS_VISIBLE | WS_CHILD,
                     10, 10, 120, 20, hwnd, NULL, NULL, NULL);

        // Ô nhập đường dẫn thư mục dự án (chỉ đọc)
        g_hwndDirPath = CreateWindow(L"EDIT", L"", WS_VISIBLE | WS_CHILD | WS_BORDER | ES_AUTOHSCROLL | ES_READONLY,
                                     10, 30, 450, 25, hwnd, NULL, NULL, NULL);

        // Nút "Browse..." để chọn thư mục dự án
        HWND hwndBrowseDir = CreateWindow(L"BUTTON", L"Browse...", WS_VISIBLE | WS_CHILD | BS_PUSHBUTTON,
                                          470, 30, 80, 25, hwnd, (HMENU)1, NULL, NULL);

        // Label "Output Directory:"
        CreateWindow(L"STATIC", L"Output Directory:", WS_VISIBLE | WS_CHILD,
                     10, 65, 120, 20, hwnd, NULL, NULL, NULL);

        // Ô nhập đường dẫn thư mục xuất file (chỉ đọc)
        g_hwndFilePath = CreateWindow(L"EDIT", L"", WS_VISIBLE | WS_CHILD | WS_BORDER | ES_AUTOHSCROLL | ES_READONLY,
                                      10, 85, 450, 25, hwnd, NULL, NULL, NULL);

        // Nút "Browse..." để chọn thư mục xuất file
        HWND hwndBrowseFile = CreateWindow(L"BUTTON", L"Browse...", WS_VISIBLE | WS_CHILD | BS_PUSHBUTTON,
                                           470, 85, 80, 25, hwnd, (HMENU)2, NULL, NULL);

        // Label hướng dẫn nhập phần mở rộng file
        CreateWindow(L"STATIC", L"File Extensions (comma separated, leave empty for all files):", WS_VISIBLE | WS_CHILD,
                     10, 120, 400, 20, hwnd, NULL, NULL, NULL);

        // Ô nhập danh sách phần mở rộng file
        g_hwndExtensions = CreateWindow(L"EDIT", L"go,cpp,h,txt,ipynb,py", WS_VISIBLE | WS_CHILD | WS_BORDER | ES_AUTOHSCROLL,
                                        10, 140, 540, 25, hwnd, NULL, NULL, NULL);

        // Nút "Process Directory" để bắt đầu xử lý
        HWND hwndProcess = CreateWindow(L"BUTTON", L"Process Directory", WS_VISIBLE | WS_CHILD | BS_PUSHBUTTON,
                                        200, 180, 150, 30, hwnd, (HMENU)3, NULL, NULL);

        // Tạo thanh trạng thái ở dưới cùng của cửa sổ
        g_hwndStatus = CreateWindowW(STATUSCLASSNAMEW, NULL, WS_CHILD | WS_VISIBLE,
                                     0, 0, 0, 0, hwnd, NULL, NULL, NULL);

        // Hiển thị thông báo sẵn sàng trên thanh trạng thái
        SendMessage(g_hwndStatus, SB_SETTEXT, 0, (LPARAM)L"Ready. Select a directory and output file.");

        // Căn giữa cửa sổ trên màn hình
        RECT rc;
        GetWindowRect(hwnd, &rc);
        int xPos = (GetSystemMetrics(SM_CXSCREEN) - (rc.right - rc.left)) / 2;
        int yPos = (GetSystemMetrics(SM_CYSCREEN) - (rc.bottom - rc.top)) / 2;
        SetWindowPos(hwnd, NULL, xPos, yPos, 0, 0, SWP_NOZORDER | SWP_NOSIZE);

        break;
    }
    case WM_SIZE: // Khi cửa sổ thay đổi kích thước
    {
        // Điều chỉnh kích thước của thanh trạng thái
        SendMessage(g_hwndStatus, WM_SIZE, 0, 0);
        break;
    }
    case WM_COMMAND: // Khi có sự kiện từ control (như nút bấm)
    {
        if (LOWORD(wParam) == 1) // Nút "Browse..." cho thư mục dự án
        {
            if (BrowseForFolder(hwnd, g_projectDir))
            {
                // Hiển thị đường dẫn đã chọn trong ô nhập liệu
                SetWindowText(g_hwndDirPath, StringToWideString(g_projectDir).c_str());
            }
        }
        else if (LOWORD(wParam) == 2) // Nút "Browse..." cho thư mục xuất file
        {
            if (BrowseForOutputDirectory(hwnd, g_outputFile))
            {
                // Hiển thị đường dẫn đã chọn trong ô nhập liệu
                SetWindowText(g_hwndFilePath, StringToWideString(g_outputFile).c_str());
            }
        }
        else if (LOWORD(wParam) == 3) // Nút "Process Directory"
        {
            // Kiểm tra xem người dùng đã chọn thư mục dự án chưa
            if (g_projectDir.empty())
            {
                MessageBox(hwnd, L"Please select a project directory.", L"Error", MB_ICONERROR);
                return 0;
            }
            // Kiểm tra xem người dùng đã chọn thư mục xuất file chưa
            if (g_outputFile.empty())
            {
                MessageBox(hwnd, L"Please select an output directory.", L"Error", MB_ICONERROR);
                return 0;
            }

            // Tạo một thread mới để xử lý thư mục mà không làm đứng giao diện
            DWORD threadId;
            CreateThread(NULL, 0, ProcessDirectoryThread, NULL, 0, &threadId);
        }
        break;
    }
    case WM_CLOSE: // Khi người dùng đóng cửa sổ
        DestroyWindow(hwnd);
        break;
    case WM_DESTROY:        // Khi cửa sổ bị hủy
        PostQuitMessage(0); // Gửi thông báo để thoát vòng lặp message
        break;
    default:
        return DefWindowProc(hwnd, msg, wParam, lParam); // Xử lý mặc định cho các message khác
    }
    return 0;
}

// Hàm entry point của ứng dụng Windows
int WINAPI wWinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, LPWSTR lpCmdLine, int nCmdShow)
{
    // Khởi tạo COM (Component Object Model) - cần thiết cho các hộp thoại chọn thư mục
    CoInitializeEx(NULL, COINIT_APARTMENTTHREADED);

    // Đăng ký lớp cửa sổ
    WNDCLASSEX wc = {0};
    wc.cbSize = sizeof(WNDCLASSEX);
    wc.style = 0;
    wc.lpfnWndProc = WndProc; // Hàm xử lý message
    wc.cbClsExtra = 0;
    wc.cbWndExtra = 0;
    wc.hInstance = hInstance;
    wc.hIcon = LoadIcon(NULL, IDI_APPLICATION);    // Biểu tượng cửa sổ
    wc.hCursor = LoadCursor(NULL, IDC_ARROW);      // Con trỏ chuột
    wc.hbrBackground = (HBRUSH)(COLOR_WINDOW + 1); // Màu nền
    wc.lpszMenuName = NULL;
    wc.lpszClassName = L"ProjectExporterClass";   // Tên lớp cửa sổ
    wc.hIconSm = LoadIcon(NULL, IDI_APPLICATION); // Biểu tượng nhỏ

    // Kiểm tra xem việc đăng ký lớp cửa sổ có thành công không
    if (!RegisterClassEx(&wc))
    {
        MessageBox(NULL, L"Window Registration Failed!", L"Error", MB_ICONEXCLAMATION | MB_OK);
        return 0;
    }

    // Tạo cửa sổ chính
    g_hwndMain = CreateWindowEx(
        WS_EX_CLIENTEDGE,                       // Kiểu mở rộng
        L"ProjectExporterClass",                // Tên lớp đã đăng ký
        L"Codebase Exporter",                   // Tiêu đề cửa sổ
        WS_OVERLAPPEDWINDOW,                    // Kiểu cửa sổ
        CW_USEDEFAULT, CW_USEDEFAULT, 570, 280, // Vị trí và kích thước
        NULL, NULL, hInstance, NULL);           // Các tham số khác

    // Kiểm tra xem việc tạo cửa sổ có thành công không
    if (g_hwndMain == NULL)
    {
        MessageBox(NULL, L"Window Creation Failed!", L"Error", MB_ICONEXCLAMATION | MB_OK);
        return 0;
    }

    // Hiển thị cửa sổ
    ShowWindow(g_hwndMain, nCmdShow);
    UpdateWindow(g_hwndMain);

    // Vòng lặp message - xử lý các sự kiện của cửa sổ
    MSG msg;
    while (GetMessage(&msg, NULL, 0, 0))
    {
        TranslateMessage(&msg); // Chuyển đổi message bàn phím
        DispatchMessage(&msg);  // Gửi message đến WndProc
    }

    // Giải phóng COM khi kết thúc
    CoUninitialize();
    return (int)msg.wParam;
}

// Hàm tạo cấu trúc thư mục dạng cây trong file output
void build_tree_structure(const fs::path &path, std::ostream &out, std::string prefix)
{
    try
    {
        // Thu thập tất cả các mục trong thư mục
        std::vector<fs::directory_entry> entries;
        for (const auto &entry : fs::directory_iterator(path))
        {
            entries.push_back(entry);
        }

        // Duyệt qua từng mục và tạo cấu trúc cây
        for (size_t i = 0; i < entries.size(); ++i)
        {
            const auto &entry = entries[i];
            // Ký tự kết nối cho cây thư mục (└── cho mục cuối cùng, ├── cho các mục khác)
            std::string connector = (i == entries.size() - 1) ? "└── " : "├── ";
            out << prefix << connector << entry.path().filename().string() << "\n";

            // Nếu mục là thư mục, đệ quy để hiển thị nội dung bên trong
            if (entry.is_directory())
            {
                // Tiền tố mới cho các mục con (thêm dấu │ hoặc khoảng trắng)
                std::string new_prefix = prefix + ((i == entries.size() - 1) ? "    " : "│   ");
                build_tree_structure(entry.path(), out, new_prefix);
            }
        }
    }
    catch (const std::filesystem::filesystem_error &e)
    {
        // Xử lý lỗi khi không thể truy cập thư mục
        out << prefix << "Error accessing path: " << e.what() << "\n";
    }
}

// Hàm ghi nội dung file vào file output
void dump_file(const fs::path &filepath, std::ostream &out)
{
    // Kiểm tra xem có phải file ipynb không
    if (filepath.extension() == ".ipynb")
    {
        process_jupyter_notebook(filepath, out);
        return;
    }

    // Mở file để đọc
    std::ifstream file(filepath);
    if (!file.is_open())
    {
        // Nếu không mở được file, ghi thông báo lỗi
        out << "\n// File: " << filepath.string() << " (could not be opened)\n";
        return;
    }

    // Ghi tên file vào file output
    out << "\n// File: " << filepath.string() << "\n";
    // Đọc và ghi từng dòng của file
    std::string line;
    while (std::getline(file, line))
    {
        out << line << "\n";
    }
    file.close();
}

// Hàm kiểm tra xem có nên đọc file dựa trên phần mở rộng không
bool should_read_file(const fs::path &filepath, const std::set<std::string> &extensions)
{
    // Nếu không có phần mở rộng nào được chỉ định, đọc tất cả các file
    if (extensions.empty())
        return true;

    // Lấy phần mở rộng của file (loại bỏ dấu chấm ở đầu)
    std::string ext = filepath.extension().string();
    if (!ext.empty() && ext[0] == '.')
        ext = ext.substr(1);

    // Kiểm tra xem phần mở rộng có trong danh sách không
    return extensions.find(ext) != extensions.end();
}

// Hàm chính để xử lý thư mục và tạo file output
void process_directory(const fs::path &dir_path, const std::string &output_file, const std::set<std::string> &extensions)
{
    // Mở file output để ghi
    std::ofstream out(output_file);
    if (!out.is_open())
    {
        throw std::runtime_error("Cannot create output file.");
    }

    // Ghi các tiêu đề và thông tin tổng quan
    out << "--- Project Overview ---\n\n\n";
    out << "--- Notes ---\nThis is the complete project code. Please read and understand thoroughly.\n\n";

    // Ghi danh sách các phần mở rộng file được bao gồm
    out << "--- File Types Included ---\n";
    if (extensions.empty())
        out << "All files\n";
    else
    {
        for (const auto &ext : extensions)
            out << "." << ext << "\n";
    }
    out << "\n";

    // Ghi cấu trúc thư mục
    out << "--- Directory Structure ---\n";
    build_tree_structure(dir_path, out);

    // Ghi nội dung chi tiết của từng file
    out << "\n--- Source Code Details ---\n";
    try
    {
        // Duyệt qua tất cả các file trong thư mục và các thư mục con
        auto options = fs::directory_options::skip_permission_denied;
        for (auto &p : fs::recursive_directory_iterator(dir_path, options))
        {
            // Nếu là file thông thường và phần mở rộng phù hợp
            if (p.is_regular_file() && should_read_file(p.path(), extensions))
            {
                // Ghi nội dung file vào file output
                dump_file(p.path(), out);
            }
        }
    }
    catch (const std::filesystem::filesystem_error &e)
    {
        // Xử lý lỗi khi duyệt thư mục
        out << "Error during directory traversal: " << e.what() << "\n";
    }

    // Đóng file output
    out.close();
}

// Hàm tách chuỗi thành các token dựa trên ký tự phân cách
std::vector<std::string> split(const std::string &s, char delimiter)
{
    std::vector<std::string> tokens;
    std::string token;
    std::istringstream tokenStream(s);
    // Đọc từng token từ chuỗi, phân cách bởi delimiter
    while (std::getline(tokenStream, token, delimiter))
    {
        // Chỉ thêm các token không rỗng vào kết quả
        if (!token.empty())
            tokens.push_back(token);
    }
    return tokens;
}