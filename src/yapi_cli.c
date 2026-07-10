#include "yapi_inc.h"
#include <stdarg.h>
#include <stdio.h>
#include <string.h>

#ifdef _WIN32
#include <windows.h>
#else
#include <errno.h>
#include <pthread.h>
#include <sys/time.h>
#include <dlfcn.h>
#endif
#ifdef __linux
#include <unistd.h>
#include <sys/syscall.h>
#endif

// macro to simplify code for loading each symbol
#define YAPI_LOAD_SYMBOL(symbolName, symbol)                                       \
    if (!symbol && yapiLoadSymbol(symbolName, (YapiPointer*)&symbol, error) < 0) { \
        return YAPI_ERROR;                                                         \
    }

#define YAPI_CHECK_CLI_RETURN() \
    if (ret != YAPI_SUCCESS) {  \
        yapiGetCliError(error); \
    }                           \
    return ret;

#define YAPI_CHECK_CLI_RETURN_WITH_INFO() \
    if (ret == YAPI_ERROR) {              \
        yapiGetCliError(error);           \
    }                                     \
    return ret;

static YapiSymbols yapiSymbols = {NULL};
static void*       yapiLibHandle = NULL;

#ifdef _WIN32

static YapiResult yapiGetWindowsError(DWORD errNum, YapiErrorMsg* error, char* errMsgPrefix)
{
    char*    fallbackErrorFormat = "failed to get message for Windows Error %d";
    wchar_t* errBuf = NULL;
    DWORD    length = 0;

    DWORD status =
        FormatMessageW(FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS | FORMAT_MESSAGE_ALLOCATE_BUFFER,
                       NULL, errNum, MAKELANGID(LANG_ENGLISH, SUBLANG_ENGLISH_US), (LPWSTR)&errBuf, 0, NULL);
    if (!status && GetLastError() == ERROR_MUI_FILE_NOT_FOUND)
        FormatMessageW(FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS | FORMAT_MESSAGE_ALLOCATE_BUFFER,
                       NULL, errNum, MAKELANGID(LANG_NEUTRAL, SUBLANG_DEFAULT), (LPWSTR)&errBuf, 0, NULL);

    if (errBuf == NULL) {
        return YAPI_SUCCESS;
    }

    // strip trailing period and carriage return from message, if needed
    length = (DWORD)wcslen(errBuf);
    errBuf[length] = L'\0';

    // convert to UTF-8 encoding
    if (length > 0) {
        strcpy_s(error->buf->message, T2S_BUFFER_SIZE, errMsgPrefix);
        length = WideCharToMultiByte(CP_UTF8, 0, errBuf, -1, error->buf->message + strlen(errMsgPrefix),
                                     T2S_BUFFER_SIZE, NULL, NULL);
    }
    LocalFree(errBuf);
    return YAPI_SUCCESS;
}

YapiResult yapiOpenDynamicLib(char* yacliLibName, YapiPointer* handler, YapiErrorMsg* error)
{
    *handler = LoadLibraryA(yacliLibName);
    if (*handler != NULL) {
        yapiLibHandle = *handler;
        return YAPI_SUCCESS;
    }

    char* userProfile = getenv("USERPROFILE");
    if (userProfile != NULL) {
        char basePath[FILENAME_MAX];
        snprintf(basePath, sizeof(basePath), "%s\\.yashandb\\client\\lib", userProfile);
        char thirdPath[FILENAME_MAX];
        snprintf(thirdPath, sizeof(thirdPath), "%s\\libssl-1_1-x64.dll", basePath);
        (void)LoadLibraryA(thirdPath);
        snprintf(thirdPath, sizeof(thirdPath), "%s\\libcrypto-1_1-x64.dll", basePath);
        (void)LoadLibraryA(thirdPath);
        snprintf(thirdPath, sizeof(thirdPath), "%s\\yas_infra.dll", basePath);
        (void)LoadLibraryA(thirdPath);
        snprintf(thirdPath, sizeof(thirdPath), "%s\\%s", basePath, yacliLibName);
        *handler = LoadLibraryA(thirdPath);
        if (*handler != NULL) {
            yapiLibHandle = *handler;
            return YAPI_SUCCESS;
        }
    }

    DWORD errNum = GetLastError();
    YAPI_CALL(yapiGetWindowsError(errNum, error, "load yacli library error: "));
    return YAPI_ERROR;
}


YapiResult yapiCloseDynamicLib(YapiPointer* handler, YapiErrorMsg* error)
{
    BOOL ret = FreeLibrary(*handler);
    if (ret) {
        *handler = NULL;
        return YAPI_SUCCESS;
    }
    DWORD errNum = GetLastError();
    yapiSetError(error, errNum, "");
    return YAPI_ERROR;
}

static int yapiLoadSymbol(const char* symbolName, void** symbol, YapiErrorMsg* error)
{
    *symbol = GetProcAddress(yapiLibHandle, symbolName);
    if (*symbol != NULL) {
        return YAPI_SUCCESS;
    }
    yapiSetError(error, YAPI_ERR_LOAD_SYMBOL, "symbol %s not found in yacli library", symbolName);
    return YAPI_ERROR;
}

#else

YapiResult yapiOpenDynamicLib(char* libName, YapiPointer* handler, YapiErrorMsg* error)
{
    *handler = dlopen(libName, RTLD_LAZY);
    if (*handler) {
        yapiLibHandle = *handler;
        return YAPI_SUCCESS;
    }

    char* homeDir = getenv("HOME");
    if (homeDir != NULL) {
        char basePath[FILENAME_MAX];
        snprintf(basePath, sizeof(basePath), "%s/.yashandb/client/lib/", homeDir);
        char thirdPath[FILENAME_MAX];
        snprintf(thirdPath, sizeof(thirdPath), "%s/libcrypto.so", basePath);
        dlopen(thirdPath, RTLD_LAZY);
        snprintf(thirdPath, sizeof(thirdPath), "%s/libssl.so", basePath);
        dlopen(thirdPath, RTLD_LAZY);
        snprintf(thirdPath, sizeof(thirdPath), "%s/libyas_infra.so", basePath);
        dlopen(thirdPath, RTLD_LAZY);
        snprintf(thirdPath, sizeof(thirdPath), "%s/%s", basePath, libName);
        *handler = dlopen(thirdPath, RTLD_LAZY);
        if (*handler) {
            yapiLibHandle = *handler;
            return YAPI_SUCCESS;
        }
    }

    char* errMsg = dlerror();
    yapiSetError(error, YAPI_ERR_LOAD_SYMBOL, "load yacli library error [%s]", errMsg);
    return YAPI_ERROR;
}

YapiResult yapiCloseDynamicLib(YapiPointer* handler, YapiErrorMsg* error)
{
    int32_t ret = dlclose(*handler);
    if (ret != YAPI_SUCCESS) {
        char* errMsg = dlerror();
        yapiSetError(error, errno, errMsg);
        return YAPI_ERROR;
    }
    *handler = NULL;
    return YAPI_SUCCESS;
}

static int yapiLoadSymbol(const char* symbolName, void** symbol, YapiErrorMsg* error)
{
    *symbol = dlsym(yapiLibHandle, symbolName);
    if (!*symbol) {
        yapiSetError(error, YAPI_ERR_LOAD_SYMBOL, "symbol %s not found in yacli library", symbolName);
        return YAPI_ERROR;
    }
    return YAPI_SUCCESS;
}
#endif

YapiResult yapiCliAllocHandle(YapiHandleType type, YacHandle input, YacHandle* output, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacAllocHandle", yapiSymbols.fnAllocHandle)
    ret = (YapiResult)(*yapiSymbols.fnAllocHandle)(type, input, output);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliFreeHandle(YapiHandleType type, YacHandle handle, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacFreeHandle", yapiSymbols.fnHandleFree)
    ret = (YapiResult)(*yapiSymbols.fnHandleFree)(type, handle);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliGetVersion(char** version, YapiErrorMsg* error)
{
    YAPI_LOAD_SYMBOL("yacGetVersion", yapiSymbols.fnGetVersion)
    *version = (*yapiSymbols.fnGetVersion)();
    return YAPI_SUCCESS;
}

YapiResult yapiCliGetLastError(YapiErrorMsg* error)
{
    char* msg;
    char* stat;
    YAPI_LOAD_SYMBOL("yacGetLastError", yapiSymbols.fnGetLastError)
    (*yapiSymbols.fnGetLastError)(&error->buf->code, &msg, &stat, &error->buf->pos);
    strcpy(error->buf->message, msg);
    strcpy(error->buf->sqlState, stat);
    return YAPI_SUCCESS;
}

YapiResult yapiCliSetEnvAttr(YapiEnv* hEnv, YapiEnvAttr attr, void* value, int32_t length, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacSetEnvAttr", yapiSymbols.fnSetEnvAttr)
    ret = (YapiResult)(*yapiSymbols.fnSetEnvAttr)(hEnv, attr, value, length);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliGetEnvAttr(YacHandle hEnv, YapiEnvAttr attr, void* value, int32_t bufLength, int32_t* stringLength,
                             YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacGetEnvAttr", yapiSymbols.fnGetEnvAttr)
    ret = (YapiResult)(*yapiSymbols.fnGetEnvAttr)(hEnv, attr, value, bufLength, stringLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliConnect(YacHandle hConn, const char* url, int16_t urlLength, const char* user, int16_t userLength,
                          const char* password, int16_t passwordLength, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacConnect", yapiSymbols.fnConnect)
    ret = (YapiResult)(*yapiSymbols.fnConnect)(hConn, url, urlLength, user, userLength, password, passwordLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDisconnect(YacHandle hConn, YapiErrorMsg* error)
{
    YAPI_LOAD_SYMBOL("yacDisconnect", yapiSymbols.fnDisconnect)
    (*yapiSymbols.fnDisconnect)(hConn);
    return YAPI_SUCCESS;
}

YapiResult yapiCliSetConnAttr(YacHandle hConn, YapiConnAttr attr, void* value, int32_t length, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacSetConnAttr", yapiSymbols.fnSetConnAttr)
    ret = (YapiResult)(*yapiSymbols.fnSetConnAttr)(hConn, attr, value, length);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliGetConnAttr(YacHandle hConn, YapiConnAttr attr, void* value, int32_t bufLength, int32_t* stringLength,
                              YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacGetConnAttr", yapiSymbols.fnGetConnAttr)
    ret = (YapiResult)(*yapiSymbols.fnGetConnAttr)(hConn, attr, value, bufLength, stringLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliCommit(YacHandle hConn, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacCommit", yapiSymbols.fnCommit)
    ret = (YapiResult)(*yapiSymbols.fnCommit)(hConn);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliRollback(YacHandle hConn, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacRollback", yapiSymbols.fnRollback)
    ret = (YapiResult)(*yapiSymbols.fnRollback)(hConn);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliCancel(YacHandle hConn, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacCancel", yapiSymbols.fnCancel)
    ret = (YapiResult)(*yapiSymbols.fnCancel)(hConn);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliPing(YacHandle hConn, int32_t timeout, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPingWithTimeout", yapiSymbols.fnPing)
    ret = (YapiResult)(*yapiSymbols.fnPing)(hConn, timeout);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliParseSqlParams(YacHandle hEnv, YacHandle* paramList, const char* sql, int32_t sqlLength, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacParseSqlParams", yapiSymbols.fnParseSqlParams)
    ret = (YapiResult)(*yapiSymbols.fnParseSqlParams)(hEnv, paramList, sql, sqlLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliGetParamListCount(YacHandle hParamList, uint32_t* count, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacGetParamListCount", yapiSymbols.fnGetParamListCount)
    ret = (YapiResult)(*yapiSymbols.fnGetParamListCount)(hParamList, count);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliGetParamName(YacHandle hParamList, uint16_t index, char* name, int32_t nameBufLen, int32_t* nameLen, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacGetParamName", yapiSymbols.fnGetParamName)
    ret = (YapiResult)(*yapiSymbols.fnGetParamName)(hParamList, index, name, nameBufLen, nameLen);
    YAPI_CHECK_CLI_RETURN();
}   

YapiResult yapiCliFreeParamList(YacHandle hParamList, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacFreeParamList", yapiSymbols.fnFreeParamList)
    ret = (YapiResult)(*yapiSymbols.fnFreeParamList)(hParamList);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliGetSqlParamCount(const char* sql, int32_t sqlLength, uint16_t* paramCount, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacGetSqlParamCount", yapiSymbols.fnGetSqlParamCount)
    ret = (YapiResult) (*yapiSymbols.fnGetSqlParamCount)(sql, sqlLength, paramCount);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDirectExecute(YacHandle hStmt, const char* sql, int32_t sqlLength, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDirectExecute", yapiSymbols.fnDirectExecute)
    ret = (YapiResult)(*yapiSymbols.fnDirectExecute)(hStmt, sql, sqlLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliPrepare(YacHandle hStmt, const char* sql, int32_t sqlLength, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPrepare", yapiSymbols.fnPrepare)
    ret = (YapiResult)(*yapiSymbols.fnPrepare)(hStmt, sql, sqlLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliExecute(YacHandle hStmt, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacExecute", yapiSymbols.fnExecute)
    ret = (YapiResult)(*yapiSymbols.fnExecute)(hStmt);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliSetStmtAttr(YacHandle hStmt, YapiStmtAttr attr, void* value, int32_t length, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacSetStmtAttr", yapiSymbols.fnSetStmtAttr)
    ret = (YapiResult)(*yapiSymbols.fnSetStmtAttr)(hStmt, attr, value, length);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliGetStmtAttr(YacHandle hStmt, YapiStmtAttr attr, void* value, int32_t bufLength, int32_t* stringLength,
                              YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacGetStmtAttr", yapiSymbols.fnGetStmtAttr)
    ret = (YapiResult)(*yapiSymbols.fnGetStmtAttr)(hStmt, attr, value, bufLength, stringLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliFetch(YacHandle hStmt, uint32_t* rows, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacFetch", yapiSymbols.fnFetch)
    ret = (YapiResult)(*yapiSymbols.fnFetch)(hStmt, rows);
    YAPI_CHECK_CLI_RETURN_WITH_INFO();
}

YapiResult yapiCliDescribeCol2(YacHandle hStmt, uint16_t id, YapiColumnDesc* desc, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDescribeCol2", yapiSymbols.fnDescribeCol2)
    ret = (YapiResult)(*yapiSymbols.fnDescribeCol2)(hStmt, id, desc);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliBindColumn(YacHandle hStmt, uint16_t id, YapiType type, YapiPointer value, int32_t bufLen,
                             int32_t* indicator, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacBindColumn", yapiSymbols.fnBindColumn)
    ret = (YapiResult)(*yapiSymbols.fnBindColumn)(hStmt, id, type, value, bufLen, indicator);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliBindParameter(YacHandle hStmt, uint16_t id, YapiParamDirection direction, YapiType bindType,
                                YapiPointer value, int32_t bindSize, int32_t bufLength, int32_t* indicator,
                                YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacBindParameter", yapiSymbols.fnBindParameter)
    ret = (YapiResult)(*yapiSymbols.fnBindParameter)(hStmt, id, direction, bindType, value, bindSize, bufLength,
                                                     indicator);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliBindParameterByName(YacHandle hStmt, char* name, YapiParamDirection direction, YapiType bindType,
                                      YapiPointer value, int32_t bindSize, int32_t bufLength, int32_t* indicator,
                                      YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacBindParameterByName", yapiSymbols.fnBindParameterByName)
    ret = (YapiResult)(*yapiSymbols.fnBindParameterByName)(hStmt, name, direction, bindType, value, bindSize, bufLength,
                                                           indicator);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliNumResultCols(YacHandle hStmt, int16_t* count, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacNumResultCols", yapiSymbols.fnNumResultCols)
    ret = (YapiResult)(*yapiSymbols.fnNumResultCols)(hStmt, count);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliColAttribute(YacHandle hStmt, uint16_t id, YapiColAttr attr, void* value, int32_t bufLen,
                               int32_t* stringLength, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacColAttribute", yapiSymbols.fnColAttribute)
    ret = (YapiResult)(*yapiSymbols.fnColAttribute)(hStmt, id, attr, value, bufLen, stringLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliNumParams(YacHandle hStmt, int16_t* count, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacNumParams", yapiSymbols.fnNumParams)
    ret = (YapiResult)(*yapiSymbols.fnNumParams)(hStmt, count);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliGetDateStruct(YapiDate date, YapiDateStruct* ds, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacGetDateStruct", yapiSymbols.fnGetDateStruct)
    ret = (YapiResult)(*yapiSymbols.fnGetDateStruct)(date, ds);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliLobDescAlloc(YacHandle* hConn, YapiType type, void** desc, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacLobDescAlloc", yapiSymbols.fnLobDescAlloc)
    ret = (YapiResult)(*yapiSymbols.fnLobDescAlloc)(hConn, type, desc);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliLobDescFree(void* desc, YapiType type, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacLobDescFree", yapiSymbols.fnLobDescFree)
    ret = (YapiResult)(*yapiSymbols.fnLobDescFree)(desc, type);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliLobGetChunkSize(YacHandle* hConn, YapiLobLocator* locator, uint16_t* chunkSize, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacLobGetChunkSize", yapiSymbols.fnLobGetChunkSize)
    ret = (YapiResult)(*yapiSymbols.fnLobGetChunkSize)(hConn, locator, chunkSize);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliLobGetLength(YacHandle* hConn, YapiLobLocator* locator, uint64_t* length, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacLobGetLength", yapiSymbols.fnLobGetLength)
    ret = (YapiResult)(*yapiSymbols.fnLobGetLength)(hConn, locator, length);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliLobRead(YacHandle* hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf, uint64_t bufLen,
                          YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacLobRead", yapiSymbols.fnLobRead)
    ret = (YapiResult)(*yapiSymbols.fnLobRead)(hConn, loc, bytes, buf, bufLen);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliLobWrite(YacHandle* hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf, uint64_t bufLen,
                           YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacLobWrite", yapiSymbols.fnLobWrite)
    ret = (YapiResult)(*yapiSymbols.fnLobWrite)(hConn, loc, bytes, buf, bufLen);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliLobCreateTemporary(YacHandle* hConn, YapiLobLocator* loc, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacLobCreateTemporary", yapiSymbols.fnLobCreateTemporary)
    ret = (YapiResult)(*yapiSymbols.fnLobCreateTemporary)(hConn, loc);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliLobFreeTemporary(YacHandle* hConn, YapiLobLocator* loc, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacLobFreeTemporary", yapiSymbols.fnLobFreeTemporary)
    ret = (YapiResult)(*yapiSymbols.fnLobFreeTemporary)(hConn, loc);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDateGetDate(const YapiDate date, int16_t* year, uint8_t* month, uint8_t* day, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDateGetDate", yapiSymbols.fnDateGetDate)
    ret = (YapiResult)(*yapiSymbols.fnDateGetDate)(date, year, month, day);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliShortTimeGetShortTime(const YapiShortTime time, uint8_t* hour, uint8_t* minute, uint8_t* second,
                                        uint32_t* fraction, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacShortTimeGetShortTime", yapiSymbols.fnShortTimeGetShortTime)
    ret = (YapiResult)(*yapiSymbols.fnShortTimeGetShortTime)(time, hour, minute, second, fraction);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliTimestampGetTimestamp(const YapiTimestamp timestamp, int16_t* year, uint8_t* month, uint8_t* day,
                                        uint8_t* hour, uint8_t* minute, uint8_t* second, uint32_t* fraction,
                                        YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacTimestampGetTimestamp", yapiSymbols.fnTimestampGetTimestamp)
    ret = (YapiResult)(*yapiSymbols.fnTimestampGetTimestamp)(timestamp, year, month, day, hour, minute, second, fraction);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliYMIntervalGetYearMonth(const YapiYMInterval ymInterval, int32_t* year, int32_t* month,
                                         YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacYMIntervalGetYearMonth", yapiSymbols.fnYMIntervalGetYearMonth)
    ret = (YapiResult)(*yapiSymbols.fnYMIntervalGetYearMonth)(ymInterval, year, month);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDSIntervalGetDaySecond(const YapiDSInterval dsInterval, int32_t* day, int32_t* hour, int32_t* minute,
                                         int32_t* second, int32_t* fraction, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDSIntervalGetDaySecond", yapiSymbols.fnDSIntervalGetDaySecond)
    ret = (YapiResult)(*yapiSymbols.fnDSIntervalGetDaySecond)(dsInterval, day, hour, minute, second, fraction);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDateSetDate(YapiDate* date, int16_t year, uint8_t month, uint8_t day, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDateSetDate", yapiSymbols.fnDateSetDate)
    ret = (YapiResult)(*yapiSymbols.fnDateSetDate)(date, year, month, day);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliShortTimeSetShortTime(YapiShortTime* time, uint8_t hour, uint8_t minute, uint8_t second,
                                        uint32_t fraction, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacShortTimeSetShortTime", yapiSymbols.fnShortTimeSetShortTime)
    ret = (YapiResult)(*yapiSymbols.fnShortTimeSetShortTime)(time, hour, minute, second, fraction);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliTimestampSetTimestamp(YapiTimestamp* timestamp, int16_t year, uint8_t month, uint8_t day,
                                        uint8_t hour, uint8_t minute, uint8_t second, uint32_t fraction,
                                        YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacTimestampSetTimestamp", yapiSymbols.fnTimestampSetTimestamp)
    ret =(YapiResult)(*yapiSymbols.fnTimestampSetTimestamp)(timestamp, year, month, day, hour, minute, second, fraction);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliYMIntervalSetYearMonth(YapiYMInterval* ymInterval, int32_t year, int32_t month, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacYMIntervalSetYearMonth", yapiSymbols.fnYMIntervalSetYearMonth)
    ret = (YapiResult)(*yapiSymbols.fnYMIntervalSetYearMonth)(ymInterval, year, month);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDateTimeGetTimeZoneOffset(YacHandle hEnv, YapiTimestamp timestamp, int8_t* hr, int8_t* mm,
                                            YapiErrorMsg* error)
{
    YapiResult ret;
    YAPI_LOAD_SYMBOL("yacDateTimeGetTimeZoneOffset", yapiSymbols.fnDateTimeGetTimeZoneOffset)
    ret = (YapiResult)(*yapiSymbols.fnDateTimeGetTimeZoneOffset)(hEnv, timestamp, hr, mm);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDSIntervalSetDaySecond(YapiDSInterval* dsInterval, int32_t day, int32_t hour, int32_t minute,
                                         int32_t second, int32_t fraction, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDSIntervalSetDaySecond", yapiSymbols.fnDSIntervalSetDaySecond)
    ret = (YapiResult)(*yapiSymbols.fnDSIntervalSetDaySecond)(dsInterval, day, hour, minute, second, fraction);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDSIntervalFromText(YacHandle hEnv, YapiDSInterval* dsInterval, const char* str, uint32_t strLen,
                                     YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDSIntervalFromText", yapiSymbols.fnDSIntervalFromText)
    ret = (YapiResult)(*yapiSymbols.fnDSIntervalFromText)(hEnv, dsInterval, str, strLen);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliYMIntervalFromText(YacHandle hEnv, YapiYMInterval* ymInterval, const char* str, uint32_t strLen,
                                     YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacYMIntervalFromText", yapiSymbols.fnYMIntervalFromText)
    ret = (YapiResult)(*yapiSymbols.fnYMIntervalFromText)(hEnv, ymInterval, str, strLen);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliNumberRound(YapiNumber* n, int32_t precision, int32_t scale, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacNumberRound", yapiSymbols.fnNumberRound)
    ret = (YapiResult)(*yapiSymbols.fnNumberRound)(n, precision, scale);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliNumberFromText(const char* str, uint32_t strLength, const char* fmt, uint32_t fmtLength,
                                 const char* nlsParam, uint32_t nlsParamLength, YapiNumber* number, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacNumberFromText", yapiSymbols.fnNumberFromText)
    ret = (YapiResult)(*yapiSymbols.fnNumberFromText)(str, strLength, fmt, fmtLength, nlsParam, nlsParamLength, number);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliNumberToText(const YapiNumber* number, const char* fmt, uint32_t fmtLength, const char* nlsParam,
                               uint32_t nlsParamLength, char* str, int32_t bufLength, int32_t* length,
                               YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacNumberToText", yapiSymbols.fnNumberToText)
    ret = (YapiResult)(*yapiSymbols.fnNumberToText)(number, fmt, fmtLength, nlsParam, nlsParamLength, str, bufLength,
                                                    length);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliNumberFromReal(const YapiPointer rnum, uint32_t length, YapiNumber* number, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacNumberFromReal", yapiSymbols.fnNumberFromReal)
    ret = (YapiResult)(*yapiSymbols.fnNumberFromReal)(rnum, length, number);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliNumberToReal(const YapiNumber* number, uint32_t length, YapiPointer rsl, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacNumberToReal", yapiSymbols.fnNumberToReal)
    ret = (YapiResult)(*yapiSymbols.fnNumberToReal)(number, length, rsl);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgStart(YacHandle hStmt, uint64_t objId, uint16_t subId, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgStart", yapiSymbols.fnPdbgStart)
    ret = (YapiResult)(*yapiSymbols.fnPdbgStart)(hStmt, objId, subId);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgCheckVersion(YacHandle hStmt, uint64_t objId, uint16_t subId, uint32_t version,
                                  YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgCheckVersion", yapiSymbols.fnPdbgCheckVersion)
    ret = (YapiResult)(*yapiSymbols.fnPdbgCheckVersion)(hStmt, objId, subId, version);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgAbort(YacHandle hStmt, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgAbort", yapiSymbols.fnPdbgAbort)
    ret = (YapiResult)(*yapiSymbols.fnPdbgAbort)(hStmt);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgContinue(YacHandle hStmt, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgContinue", yapiSymbols.fnPdbgContinue)
    ret = (YapiResult)(*yapiSymbols.fnPdbgContinue)(hStmt);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgStepInto(YacHandle hStmt, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgStepInto", yapiSymbols.fnPdbgStepInto)
    ret = (YapiResult)(*yapiSymbols.fnPdbgStepInto)(hStmt);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgStepOut(YacHandle hStmt, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgStepOut", yapiSymbols.fnPdbgStepOut)
    ret = (YapiResult)(*yapiSymbols.fnPdbgStepOut)(hStmt);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgStepNext(YacHandle hStmt, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgStepNext", yapiSymbols.fnPdbgStepNext)
    ret = (YapiResult)(*yapiSymbols.fnPdbgStepNext)(hStmt);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgDeleteAllBreakpoints(YacHandle hStmt, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgDeleteAllBreakpoints", yapiSymbols.fnPdbgDeleteAllBreakpoints)
    ret = (YapiResult)(*yapiSymbols.fnPdbgDeleteAllBreakpoints)(hStmt);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgAddBreakpoint(YacHandle hStmt, uint64_t objId, uint16_t subId, uint32_t lineNo, uint32_t* bpId,
                                   YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgAddBreakpoint", yapiSymbols.fnPdbgAddBreakpoint)
    ret = (YapiResult)(*yapiSymbols.fnPdbgAddBreakpoint)(hStmt, objId, subId, lineNo, bpId);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgGetBreakpointsCount(YacHandle hStmt, uint32_t* bpCount, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetBreakpointsCount", yapiSymbols.fnPdbgGetBreakpointsCount)
    ret = (YapiResult)(*yapiSymbols.fnPdbgGetBreakpointsCount)(hStmt, bpCount);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgDeleteBreakpoint(YacHandle hStmt, uint64_t objId, uint16_t subId, uint32_t lineNo,
                                      YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgDeleteBreakpoint", yapiSymbols.fnPdbgDeleteBreakpoint)
    ret = (YapiResult)(*yapiSymbols.fnPdbgDeleteBreakpoint)(hStmt, objId, subId, lineNo);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgGetAllVars(YacHandle hStmt, uint32_t* varCount, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetAllVars", yapiSymbols.fnPdbgPdbgGetAllVars)
    ret = (YapiResult)(*yapiSymbols.fnPdbgPdbgGetAllVars)(hStmt, varCount);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgGetAllFrames(YacHandle hStmt, uint32_t* frameCount, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetAllFrames", yapiSymbols.fnPdbgGetAllFrames)
    ret = (YapiResult)(*yapiSymbols.fnPdbgGetAllFrames)(hStmt, frameCount);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgGetRunningAttrs(YacHandle hStmt, YapiDebugRunningAttr attr, void* value, int32_t bufLen,
                                     int32_t* stringLength, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetRunningAttrs", yapiSymbols.fnPdbgGetRunningAttrs)
    ret = (YapiResult)(*yapiSymbols.fnPdbgGetRunningAttrs)(hStmt, attr, value, bufLen, stringLength);
    YAPI_CHECK_CLI_RETURN();
}
YapiResult yapiCiPdbgGetFrameAttrs(YacHandle hStmt, uint32_t id, YapiDebugFrameAttr attr, void* value, int32_t bufLen,
                                   int32_t* stringLength, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetFrameAttrs", yapiSymbols.fnPdbgGetFrameAttrs)
    ret = (YapiResult)(*yapiSymbols.fnPdbgGetFrameAttrs)(hStmt, id, attr, value, bufLen, stringLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgGetVarAttrs(YacHandle hStmt, uint32_t id, YapiDebugVarAttr attr, void* value, int32_t bufLen,
                                 int32_t* stringLength, YapiErrorMsg* error)

{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetVarAttrs", yapiSymbols.fnPdbgGetVarAttrs)
    ret = (YapiResult)(*yapiSymbols.fnPdbgGetVarAttrs)(hStmt, id, attr, value, bufLen, stringLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgGetVarValue(YacHandle hStmt, uint32_t id, uint32_t valueType, void* value, int32_t bufLen,
                                 int32_t* indicator, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetVarValue", yapiSymbols.fnPdbgGetVarValue)
    ret = (YapiResult)(*yapiSymbols.fnPdbgGetVarValue)(hStmt, id, valueType, value, bufLen, indicator);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgGetBreakpointAttrs(YacHandle hStmt, uint32_t id, YapiDebugBpAttr attr, void* value, int32_t bufLen,
                                        int32_t* stringLength, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetBreakpointAttrs", yapiSymbols.fnPdbgGetBreakpointAttrs)
    ret = (YapiResult)(*yapiSymbols.fnPdbgGetBreakpointAttrs)(hStmt, id, attr, value, bufLen, stringLength);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCiPdbgGetOutput(YacHandle hStmt, char* buffer, uint32_t* len, bool* hasMore, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacPdbgGetOutput", yapiSymbols.fnPdbgGetOutput)
    ret = (YapiResult)(*yapiSymbols.fnPdbgGetOutput)(hStmt, buffer, len, hasMore);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliConnectionPoolCreate(YacHandle hConnPool, const char* url, int16_t urlLength,
                                       uint32_t min, uint32_t max, uint32_t increment, const char* user, int16_t userLength,
                                       const char* password, int16_t passwordLength, uint32_t mode, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacConnectionPoolCreate", yapiSymbols.fnConnectionPoolCreate)
    ret = (YapiResult)(*yapiSymbols.fnConnectionPoolCreate)(hConnPool, url, urlLength, min, max, increment, user,
                                                            userLength, password, passwordLength, mode);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliConnectionGet(YacHandle hConnPool, YacHandle* hConn, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacConnectionGet", yapiSymbols.fnConnectionGet)
    ret = (YapiResult) (*yapiSymbols.fnConnectionGet)(hConnPool, hConn);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliConnectionGiveBack(YacHandle hConn, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacConnectionGiveBack", yapiSymbols.fnConnectionGiveBack)
    ret = (YapiResult) (*yapiSymbols.fnConnectionGiveBack)(hConn);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliConnectionPoolDestroy(YacHandle hConnPool, uint32_t mode, YapiErrorMsg* error)
{
    YAPI_LOAD_SYMBOL("yacConnectionPoolDestroy", yapiSymbols.fnConnectionPoolDestroy)
    return YAPI_SUCCESS;
}

YapiResult yapiCliDescAlloc2(YacHandle hEnv, void** desc, YapiDescType type, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDescAlloc2", yapiSymbols.fnDescAlloc2)
    ret = (YapiResult) (*yapiSymbols.fnDescAlloc2)(hEnv, desc, type);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliDescFree2(YacHandle hEnv, void* desc, YapiDescType type, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacDescFree2", yapiSymbols.fnDescFree2)
    ret = (YapiResult) (*yapiSymbols.fnDescFree2)(hEnv, desc, type);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliVectorFromText(YapiVector* vector, YapiVectorFormat format, uint16_t dim, char* text, uint32_t textlen, uint32_t mode, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("YapiVectorFromText", yapiSymbols.fnVectorFromText)
    ret = (YapiResult) (*yapiSymbols.fnVectorFromText)(vector, format, dim, text, textlen, mode);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliVectorFromArray(YapiVector* vector, YapiVectorFormat format, uint16_t dim, uint8_t* array, uint32_t arrayLen, uint32_t mode, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacVectorFromArray", yapiSymbols.fnVectorFromArray)
    ret = (YapiResult) (*yapiSymbols.fnVectorFromArray)(vector, format, dim, array, arrayLen, mode);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliVectorToText(YapiVector* vector, char* text, uint32_t* textlen, uint32_t mode, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacVectorToText", yapiSymbols.fnVectorToText)
    ret = (YapiResult) (*yapiSymbols.fnVectorToText)(vector, text, textlen, mode);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliVectorToArray(YapiVector* vector, YapiVectorFormat format, uint16_t* dim, uint8_t* array, uint32_t* arrayLen, uint32_t mode, YapiErrorMsg* error)
{
    YapiResult ret;

    YAPI_LOAD_SYMBOL("yacVectorToArray", yapiSymbols.fnVectorToArray)
    ret = (YapiResult) (*yapiSymbols.fnVectorToArray)(vector, format, dim, array, arrayLen, mode);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliVectorGetFormat(YapiVector* vector, YapiVectorFormat* format, YapiErrorMsg* error)
{
    YapiResult ret;
    
    YAPI_LOAD_SYMBOL("yacVectorGetFormat", yapiSymbols.fnVectorGetFormat)
    ret = (YapiResult) (*yapiSymbols.fnVectorGetFormat)(vector, format);
    YAPI_CHECK_CLI_RETURN();
}

YapiResult yapiCliVectorGetDimension(YapiVector* vector, uint16_t* dim, YapiErrorMsg* error)
{
    YapiResult ret;
    
    YAPI_LOAD_SYMBOL("yacVectorGetDimension", yapiSymbols.fnVectorGetDimension)
    ret = (YapiResult) (*yapiSymbols.fnVectorGetDimension)(vector, dim);
    YAPI_CHECK_CLI_RETURN();
}
