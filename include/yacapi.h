#ifndef YAPI_API_H
#define YAPI_API_H

#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>

#ifdef __cplusplus
extern "C" {
#endif

#define YAPI_TRUE true
#define YAPI_FALSE false

#define YAPI_PARAM_NAME_BUFFER_SIZE 32
#define YAPI_MIN_PACKET_SIZE KB(64)
#define YAPI_MAX_PACKET_SIZE MB(32)
#define YAPI_NULL_DATA -1
#define YAPI_NULL_TERM_STR -2
#define YAPI_MAX_ERROR_MSG_LEN 4096
#define YAPI_MAX_SQLSTAT_LEN 16

typedef int64_t YapiDate;
typedef int64_t YapiShortTime;
typedef int32_t YapiYMInterval;
typedef int64_t YapiDSInterval;
typedef void*   YapiPointer;

typedef struct StYapiConnect YapiConnect;
typedef struct StYapiStmt    YapiStmt;
typedef struct StYapiEnv     YapiEnv;
typedef void* YacHandle;
typedef struct StYapiConnectPool YapiConnectPool;

#pragma pack(4)
#define YAC_NUMBER_SIZE 20
typedef struct StYapiNumber {
    char numberPart[YAC_NUMBER_SIZE];
} YapiNumber;

typedef struct StYapiTimestamp {
    int64_t stamp;
    int16_t bias;  // minutes
    int16_t unused;
} YapiTimestamp;
#pragma pack()

typedef struct StYapiDateStruct {
    uint16_t year;
    uint8_t  month;
    uint8_t  day;
    uint8_t  hour;
    uint8_t  minute;
    uint8_t  second;
    uint8_t  dayOfWeek;
    uint8_t  weekName;
    uint16_t dayOfYear;
    uint8_t  unused[5];
    uint32_t fraction;
    uint32_t secondOfDay;
} YapiDateStruct;

typedef enum EnYapiType {
    YAPI_TYPE_UNKNOWN = 0,
    YAPI_TYPE_BOOL = 1,
    YAPI_TYPE_TINYINT = 2,
    YAPI_TYPE_SMALLINT = 3,
    YAPI_TYPE_INTEGER = 4,
    YAPI_TYPE_BIGINT = 5,
    YAPI_TYPE_UTINYINT = 6,
    YAPI_TYPE_USMALLINT = 7,
    YAPI_TYPE_UINTEGER = 8,
    YAPI_TYPE_UBIGINT = 9,
    YAPI_TYPE_FLOAT = 10,
    YAPI_TYPE_DOUBLE = 11,
    YAPI_TYPE_NUMBER = 12,
    YAPI_TYPE_DATE = 13,
    YAPI_TYPE_SHORTDATE = 14,
    YAPI_TYPE_SHORTTIME = 15,
    YAPI_TYPE_TIMESTAMP = 16,
    YAPI_TYPE_TIMESTAMP_LTZ = 17,
    YAPI_TYPE_TIMESTAMP_TZ = 18,
    YAPI_TYPE_YM_INTERVAL = 19,
    YAPI_TYPE_DS_INTERVAL = 20,
    // 21-23 reversed
    YAPI_TYPE_CHAR = 24,
    YAPI_TYPE_NCHAR = 25,
    YAPI_TYPE_VARCHAR = 26,
    YAPI_TYPE_NVARCHAR = 27,
    YAPI_TYPE_BINARY = 28,
    YAPI_TYPE_CLOB = 29,
    YAPI_TYPE_BLOB = 30,
    YAPI_TYPE_BIT = 31,
    YAPI_TYPE_ROWID = 32,
    YAPI_TYPE_NCLOB = 33,
    YAPI_TYPE_CURSOR = 34,
    YAPI_TYPE_JSON = 35,
    YAPI_TYPE_XML = 39,
    YAPI_TYPE_NUMBER_FLOAT = 40,
    YAPI_TYPE_BFILE = 41,
    YAPI_TYPE_VECTOR = 42,
    __YAPI_TYPES_COUNT__
} YapiType;

typedef enum EnYapiResult { YAPI_SUCCESS = 0, YAPI_SUCCESS_WITH_INFO = 1, YAPI_ERROR = -1 } YapiResult;

typedef struct StYapiTextPos {
    int32_t line;
    int32_t column;
} YapiTextPos;

typedef enum StYapiCredtType { CRED_RDBMS = 0, CRED_EXT, __YAC_CRED_COUNT__ } YapiCredtType;

// structure used for transferring error information from YAC API
typedef struct StYapiErrorInfo {
    int32_t      errCode;
    char*        message;
    char*        sqlState;
    YapiTextPos* pos;
} YapiErrorInfo;

typedef struct StYapiColumnDesc {
    char*    name;
    uint32_t size;
    uint8_t  type;
    union {
      struct {
        uint8_t precision;
        int8_t scale;
      };
      struct {
        uint8_t format : 4;
        uint8_t unused: 4;
        uint8_t reserved;
      } vector;
    };
    uint8_t  nullable;
} YapiColumnDesc;

typedef enum EnYapiDebugRunningAttr {
    YAPI_DBG_RUNNING_ATTR_STATUS = 0,
    YAPI_DBG_RUNNING_ATTR_OBJ_ID = 1,
    YAPI_DBG_RUNNING_ATTR_SUB_ID = 2,
    YAPI_DBG_RUNNING_ATTR_LINE_NO = 3,
    YAPI_DBG_RUNNING_ATTR_CLASS_NAME = 4,
    YAPI_DBG_RUNNING_ATTR_METHOD_NAME = 5,
} YapiDebugRunningAttr;

typedef enum EnYapiDebugFrameAttr {
    YAPI_DBG_FRAME_ATTR_OBJ_ID = 0,
    YAPI_DBG_FRAME_ATTR_SUB_ID = 1,
    YAPI_DBG_FRAME_ATTR_LINE_NO = 2,
    YAPI_DBG_FRAME_ATTR_STACK_NO = 3,
    YAPI_DBG_FRAME_ATTR_CLASS_NAME = 4,
    YAPI_DBG_FRAME_ATTR_METHOD_NAME = 5,
} YapiDebugFrameAttr;

typedef enum EnYapiDebugVarAttr {
    YAPI_DBG_VAR_ATTR_BLOCK_NO = 0,
    YAPI_DBG_VAR_ATTR_TYPE = 1,
    YAPI_DBG_VAR_ATTR_IS_GLOBAL = 2,
    YAPI_DBG_VAR_ATTR_NAME = 3,
    YAPI_DBG_VAR_ATTR_VALUE_SIZE = 4,
} YapiDebugVarAttr;

typedef enum EnYapiDebugBpAttr {
    YAPI_DBG_BP_ATTR_OBJ_ID = 0,
    YAPI_DBG_BP_ATTR_SUB_ID = 1,
    YAPI_DBG_BP_ATTR_LINE_NO = 2,
} YapiDebugBpAttr;

typedef enum EnYapiDebuggerStatus {
    YAPI_DBG_STATUS_OFF = 0,
    YAPI_DBG_STATUS_ON = 1,
} YapiDebuggerStatus;

typedef enum EnYapiBindType {
    YAPI_BIND_COLUMN = 0,
    YAPI_BIND_PARAM = 1,
    __YAPI_BIND_TYPE_COUNT__,
} YapiBindType;

typedef enum EnYapiParamDirection {
    YAPI_PARAM_INPUT = 1,
    YAPI_PARAM_OUTPUT = 2,
    YAPI_PARAM_INOUT = 3,
} YapiParamDirection;

typedef enum EnYapiHandleType {
    YAPI_HANDLE_UNKNOWN = 0,
    YAPI_HANDLE_ENV = 1,
    YAPI_HANDLE_DBC = 2,
    YAPI_HANDLE_STMT = 3,
    YAPI_HANDLE_DESC = 4,
    YAPI_HANDLE_PUMP = 5,
    YAPI_HANDLE_CONN_POOL = 6,
    __YAPI_HANDLE_COUNT__
} YapiHandleType;

typedef enum EnYapiCharsetCode {
    YAPI_CHARSET_ASCII = 0,
    YAPI_CHARSET_GBK,
    YAPI_CHARSET_UTF8,
    YAPI_CHARSET_ISO88591,
    __YAPI_CHARSET_END__
} YapiCharsetCode;

typedef enum EnYapiEnvAttr {
    __YAPI_ENV_ATTR_BEGIN__ = 60,
    YAPI_ATTR_DATE_FORMAT = 60,
    YAPI_ATTR_CHARSET = 61,
    YAPI_ATTR_CHARSET_CODE = 62,
    YAPI_ATTR_DATA_PATH = 63,
    YAPI_ATTR_CLIENT_DRIVER = 66,
    __YAPI_ENV_ATTR_END__
} YapiEnvAttr;

typedef enum EnYapiConnAttr {
    __YAPI_CONN_ATTR_BEGIN__ = 1,
    YAPI_ATTR_LOGONINFO_PTR = 1,
    YAPI_ATTR_ASYNC_ENABLE = 2,
    YAPI_ATTR_AUTOCOMMIT = 3,
    YAPI_ATTR_LOGIN_TIMEOUT = 4,
    YAPI_ATTR_STMTS = 5,
    YAPI_ATTR_PACKET_SIZE = 6,
    YAPI_ATTR_TXN_ISOLATION = 7,
    YAPI_ATTR_SERVEROUTPUT = 8,
    YAPI_ATTR_NUMWIDTH = 9,
    YAPI_ATTR_AUTOTRACE = 10,
    YAPI_ATTR_CREDT = 11,
    YAPI_ATTR_MAX_CHARSET_RATIO = 12,
    YAPI_ATTR_TAF_ENABLED = 14,
    YAPI_ATTR_TAF_CALLBACK = 15,
    YAPI_ATTR_MAX_NCHARSET_RATIO = 17,
    YAPI_ATTR_HEARTBEAT_ENABLED = 18,
    YAPI_ATTR_SET_COMPAT_VECTOR = 21,
    YAPI_ATTR_SERVER_STATUS = 22,
    YAPI_ATTR_ROLE = 23,
    YAPI_ATTR_SSL_ROOT_CER = 24,
    __YAPI_CONN_ATTR_END__
} YapiConnAttr;

typedef enum EnYapiStmtAttr {
    __YAPI_STMT_ATTR_BEGIN__ = 100,
    YAPI_ATTR_PARAMSET_SIZE = 100,
    YAPI_ATTR_ROWSET_SIZE,
    YAPI_ATTR_ROWS_FETECHED,
    YAPI_ATTR_ROWS_AFFECTED,
    YAPI_ATTR_CURSOR_EOF,
    YAPI_ATTR_SQLTYPE,
    YAPI_ATTR_IS_BATCHROWS,
    YAPI_ATTR_IS_BATCH_ERRORS,
    YAPI_ATTR_ACK_BATCHROWS_SIZE,
    YAPI_ATTR_ACK_BATCH_ERRORS_SIZE,
    YAPI_ATTR_ACK_BATCHROWS,
    YAPI_ATTR_ACK_BATCH_ERRORS,
    __YAPI_STMT_ATTR_END__
} YapiStmtAttr;

typedef enum EnYapiColAttr {
    __YAPI_COL_ATTR_BEGIN__ = 0,
    YAPI_COL_ATTR_DISPLAY_SIZE = 0,
    YAPI_COL_ATTR_NAME = 1,
    YAPI_COL_ATTR_SIZE = 2,
    YAPI_COL_ATTR_TYPE = 3,
    YAPI_COL_ATTR_PRECISION = 4,
    YAPI_COL_ATTR_SCALE = 5,
    YAPI_COL_ATTR_NULLABLE = 6,
    YAPI_COL_ATTR_CHAR_SIZE = 7,
    YAPI_COL_ATTR_CHAR_USED = 8,
    YAPI_COL_ATTR_DISPLAY_CHAR_SIZE = 9,
    YAPI_COL_ATTR_VECTOR_DIMENSION = 10,
    YAPI_COL_ATTR_VECTOR_DATA_FORMAT = 11,
    __YAPI_COL_ATTR_END__
} YapiColAttr;

typedef struct StYapiBatchError {
    uint32_t rowNum;
    uint32_t errCode;
    uint32_t msgLen;
    char     msg[YAPI_MAX_ERROR_MSG_LEN];
} YapiBatchError;

typedef struct StYapiLobLocator YapiLobLocator;

typedef enum EnYapiSQLType {
    YAPI_SQLTYPE_QUERY = 1,
    YAPI_SQLTYPE_INSERT,
    YAPI_SQLTYPE_UPDATE,
    YAPI_SQLTYPE_DELETE,
    YAPI_SQLTYPE_MERGE,
    YAPI_SQLTYPE_WITH,
    YAPI_SQLTYPE_ANONYMOUS_BLOCK,
    YAPI_SQLTYPE_DML_CEIL = 10,

    YAPI_SQLTYPE_CREATE_DATABASE = 11,
    YAPI_SQLTYPE_CREATE_DATASPACE,
    YAPI_SQLTYPE_CREATE_TABLESPACE_SET,
    YAPI_SQLTYPE_CREATE_TABLESPACE,
    YAPI_SQLTYPE_CREATE_TABLE,
    YAPI_SQLTYPE_CREATE_SHARDED_TABLE,
    YAPI_SQLTYPE_CREATE_DUPLICATED_TABLE,
    YAPI_SQLTYPE_CREATE_TEMP_TABLE,
    YAPI_SQLTYPE_CREATE_INDEX,
    YAPI_SQLTYPE_CREATE_AC,
    YAPI_SQLTYPE_CREATE_VIEW,
    YAPI_SQLTYPE_CREATE_SYNONYM,
    YAPI_SQLTYPE_CREATE_PROCEDURE,
    YAPI_SQLTYPE_CREATE_FUNCTION,
    YAPI_SQLTYPE_CREATE_TRIGGER,
    YAPI_SQLTYPE_CREATE_PACKAGE,
    YAPI_SQLTYPE_CREATE_OR_REPLACE,
    YAPI_SQLTYPE_CREATE_SEQUENCE,
    YAPI_SQLTYPE_CREATE_USER,
    YAPI_SQLTYPE_ALTER_DATABASE,
    YAPI_SQLTYPE_ALTER_TABLE,
    YAPI_SQLTYPE_ALTER_INDEX,
    YAPI_SQLTYPE_ALTER_TABLESPACE_SET,
    YAPI_SQLTYPE_ALTER_TABLESPACE,
    YAPI_SQLTYPE_ALTER_SEQUENCE,
    YAPI_SQLTYPE_ALTER_SYSTEM,
    YAPI_SQLTYPE_ALTER_SESSION,
    YAPI_SQLTYPE_DROP_DATASPACE,
    YAPI_SQLTYPE_DROP_TABLESPACE_SET,
    YAPI_SQLTYPE_DROP_TABLE,
    YAPI_SQLTYPE_DROP_INDEX,
    YAPI_SQLTYPE_DROP_AC,
    YAPI_SQLTYPE_DROP_SEQUENCE,
    YAPI_SQLTYPE_DROP_TABLESPACE,
    YAPI_SQLTYPE_DROP_VIEW,
    YAPI_SQLTYPE_DROP_SYNONYM,
    YAPI_SQLTYPE_DROP_USER,
    YAPI_SQLTYPE_DROP_PROCEDURE,
    YAPI_SQLTYPE_DROP_FUNCTION,
    YAPI_SQLTYPE_DROP_PACKAGE,
    YAPI_SQLTYPE_TRUNCATE_TABLE,
    YAPI_SQLTYPE_BACKUP_DATABASE,
    YAPI_SQLTYPE_RESTORE_DATABASE,
    YAPI_SQLTYPE_RECOVER_DATABASE,
    YAPI_SQLTYPE_BUILD_DATABASE,
    YAPI_SQLTYPE_SET_TRANSACTION,
    YAPI_SQLTYPE_PREPARE_TRANSACTION,
    YAPI_SQLTYPE_REPLACE_VIEW,
    YAPI_SQLTYPE_REPLACE_SYNONYM,
    YAPI_SQLTYPE_REPLACE_FUNCTION,
    YAPI_SQLTYPE_REPLACE_PROCEDURE,
    YAPI_SQLTYPE_REPLACE_PACKAGE,
    YAPI_SQLTYPE_FLASHBACK_TABLE,
    YAPI_SQLTYPE_COMMENT,
    YAPI_SQLTYPE_PURGE_RECYCLEBIN,
    YAPI_SQLTYPE_CREATE_ROLE,
    YAPI_SQLTYPE_DROP_ROLE,
    YAPI_SQLTYPE_GRANT,
    YAPI_SQLTYPE_REVOKE,
    YAPI_SQLTYPE_ALTER_USER,
    YAPI_SQLTYPE_ALTER_TRIGGER,
    YAPI_SQLTYPE_DROP_TRIGGER,
    YAPI_SQLTYPE_REPLACE_TRIGGER,
    YAPI_SQLTYPE_REPLACE_OUTLINE,
    YAPI_SQLTYPE_CREATE_AUDIT_POLICY,
    YAPI_SQLTYPE_ALTER_AUDIT_POLICY,
    YAPI_SQLTYPE_DROP_AUDIT_POLICY,
    YAPI_SQLTYPE_AUDIT_POLICY,
    YAPI_SQLTYPE_NOAUDIT_POLICY,
    YAPI_SQLTYPE_CREATE_OUTLINE,
    YAPI_SQLTYPE_ALTER_OUTLINE,
    YAPI_SQLTYPE_DROP_OUTLINE,
    YAPI_SQLTYPE_ANAYLZE_TABLE,
    YAPI_SQLTYPE_DDL_CEIL = 128,

    YAPI_SQLTYPE_COMMIT,
    YAPI_SQLTYPE_ROLLBACK,
    YAPI_SQLTYPE_EXPLAIN,
    YAPI_SQLTYPE_SAVEPOINT,
    YAPI_SQLTYPE_SHUTDOWN,
    YAPI_SQLTYPE_RELEASE_SAVEPOINT,
    YAPI_SQLTYPE_SO_ENTITY,
    YAPI_SQLTYPE_PACK_ENTITY,

    YAPI_SQLTYPE_BATCH_INSERT,
    /* distributed inner type */
    YAPI_SQLTYPE_DXG_QUERY,

    __YAPI_SQLTYPE_COUNT__ = 255
} YapiSQLType;

/* YACAPI Vector format type */
typedef enum EnYapiVectorFormat {
    YAPI_VECTOR_FORMAT_FLEX = 0,
    YAPI_VECTOR_FORMAT_FLOAT16 = 1,
    YAPI_VECTOR_FORMAT_FLOAT32 = 2,
    YAPI_VECTOR_FORMAT_FLOAT64 = 3,
    YAPI_VECTOR_FORMAT_INT8 = 4
} YapiVectorFormat;

/* YACAPI describe type codes */
typedef enum EnYapiDescType {
    YAPI_DESC_UNKNOWN = 0,
    YAPI_DESC_LOB = 1,
    YAPI_DESC_VECTOR = 2
} YapiDescType;

typedef struct StYapiVector {
    YapiEnv*    env;
    uint32_t    capacity;
    uint32_t    size;
    YapiPointer vector;
} YapiVector;

#define yapiEnvCreate yapiAllocEnv
#define yapiEnvRelease yapiReleaseEnv

#define yapiConnCreate yapiConnect
#define yapiConnDisconnect yapiDisconnect
#define yapiConnRelease yapiReleaseConn
#define yapiConnCancel yapiCancel
#define yapiConnCommit yapiCommit
#define yapiConnRollback yapiRollback
#define yapiConnSetAttr yapiSetConnAttr
#define yapiConnGetAttr yapiGetConnAttr
#define yapiConnPrepareStmt yapiPrepare

#define yapiStmtExecute yapiExecute
#define yapiStmtFetch yapiFetch
#define yapiStmtDirectExecute yapiDirectExecute
#define yapiStmtDescribeCol2 yapiDescribeCol2
#define yapiStmtBindColumn yapiBindColumn
#define yapiStmtBindParameter yapiBindParameter
#define yapiStmtBindParameterByName yapiBindParameterByName
#define yapiStmtNumResultCols yapiNumResultCols
#define yapiStmtSetAttr yapiSetStmtAttr
#define yapiStmtGetAttr yapiGetStmtAttr
#define yapiStmtRelease yapiReleaseStmt
#define yapiStmtColAttribute yapiColAttribute
#define yapiStmtNumParams yapiNumParams

char* yapiGetVersion(YapiEnv* inst);
void  yapiGetLastError(YapiErrorInfo* info);

//-----------------------------------------------------------------------------
// Enviment Function
//-----------------------------------------------------------------------------
YapiResult yapiAllocEnv(YapiEnv** inst);
YapiResult yapiReleaseEnv(YapiEnv* inst);
YapiResult yapiSetEnvAttr(YapiEnv* hEnv, YapiEnvAttr attr, void* value, int32_t length);
YapiResult yapiEnvGetAttr(YapiEnv* hEnv, YapiEnvAttr attr, void* value, int32_t bufLength, int32_t* stringLength);

//-----------------------------------------------------------------------------
// Session Function
//-----------------------------------------------------------------------------
YapiResult yapiConnect(YapiEnv* env, const char* url, int16_t urlLength, const char* user, int16_t userLength,
                       const char* password, int16_t passwordLength, YapiConnect** hConn);
YapiResult yapiDisconnect(YapiConnect* hConn);
YapiResult yapiReleaseConn(YapiConnect* hConn);
YapiResult yapiCancel(YapiConnect* hConn);
YapiResult yapiCommit(YapiConnect* hConn);
YapiResult yapiRollback(YapiConnect* hConn);
YapiResult yapiSetConnAttr(YapiConnect* hConn, YapiConnAttr attr, void* value, int32_t length);
YapiResult yapiGetConnAttr(YapiConnect* hConn, YapiConnAttr attr, void* value, int32_t bufLength,
                           int32_t* stringLength);
YapiResult yapiAllocConnect(YapiEnv* env, YapiConnect** hConn);
YapiResult yapiConnect2(YapiConnect* hConn, const char* url, int16_t urlLength, const char* user, int16_t userLength,
                        const char* password, int16_t passwordLengt);
YapiResult yapiAllocConnectionPool(YapiEnv* env, YapiConnectPool** hConnPool);
YapiResult yapiReleaseConnectionPool(YapiConnectPool* hConnPool);
YapiResult yapiConnectionPoolCreate(YapiConnectPool* hConnPool, const char* url, int16_t urlLength,
                                    uint32_t min, uint32_t max, uint32_t increment, const char* user, int16_t userLength,
                                    const char* password, int16_t passwordLength, uint32_t mode);
YapiResult yapiConnectionGet(YapiConnectPool* hConnPool, YapiConnect** hConn);
YapiResult yapiConnectionGiveBack(YapiConnect* hConn);
YapiResult yapiConnectionPoolDestroy(YapiConnectPool* hConnPool, uint32_t mode);
YapiResult yapiDescAlloc2(YapiEnv* hEnv, void** desc, YapiDescType type);
YapiResult yapiDescFree2(YapiEnv* hEnv, void* desc, YapiDescType type);
YapiResult yapiPing(YapiConnect* hConn, int32_t timeout);

//-----------------------------------------------------------------------------
// SQL Cli Parser Function
//-----------------------------------------------------------------------------
YapiResult yapiParseSqlParams(YapiEnv* hEnv, YapiPointer* paramList, const char* sql, int32_t sqlLength);
YapiResult yapiGetParamListCount(YapiPointer hParamList, uint32_t* count);
YapiResult yapiGetParamName(YapiPointer hParamList, uint16_t index, char* name, int32_t nameBufLen, int32_t* nameLen);
YapiResult yapiFreeParamList(YapiPointer hParamList);
YapiResult yapiGetSqlParamCount(const char* sql, int32_t sqlLength, uint16_t* paramCount);

//-----------------------------------------------------------------------------
// Statment Function
//-----------------------------------------------------------------------------
YapiResult yapiPrepare(YapiConnect* hConn, const char* sql, int32_t sqlLength, YapiStmt** hStmt);
YapiResult yapiStmtCreate(YapiConnect* hConn, YapiStmt** hStmt);
YapiResult yapiStmtPrepare(YapiStmt* hStmt, const char* sql, int32_t sqlLength);
YapiResult yapiExecute(YapiStmt* hStmt);
YapiResult yapiFetch(YapiStmt* hStmt, uint32_t* rows);
YapiResult yapiDirectExecute(YapiStmt* hStmt, const char* sql, int32_t sqlLength);
YapiResult yapiDescribeCol2(YapiStmt* hStmt, uint16_t id, YapiColumnDesc* desc);
YapiResult yapiBindColumn(YapiStmt* hStmt, uint16_t id, YapiType type, YapiPointer value, int32_t bufLen,
                          int32_t* indicator);
YapiResult yapiBindParameter(YapiStmt* hStmt, uint16_t id, YapiParamDirection direction, YapiType bindType,
                             YapiPointer value, int32_t bindSize, int32_t bufLength, int32_t* indicator);
YapiResult yapiBindParameterByName(YapiStmt* hStmt, char* name, YapiParamDirection direction, YapiType bindType,
                                   YapiPointer value, int32_t bindSize, int32_t bufLength, int32_t* indicator);
YapiResult yapiNumResultCols(YapiStmt* hStmt, int16_t* count);
YapiResult yapiSetStmtAttr(YapiStmt* hStmt, YapiStmtAttr attr, void* value, int32_t length);
YapiResult yapiGetStmtAttr(YapiStmt* hStmt, YapiStmtAttr attr, void* value, int32_t bufLength, int32_t* stringLength);
YapiResult yapiColAttribute(YapiStmt* hStmt, uint16_t id, YapiColAttr attr, void* value, int32_t bufLen,
                            int32_t* stringLength);
YapiResult yapiNumParams(YapiStmt* hStmt, int16_t* count);
YapiResult yapiReleaseStmt(YapiStmt* hStmt);
YapiResult yapiReleaseHandle(YacHandle* handle);
YapiResult yapiCursorStmtCreate(YapiConnect* hConn, YapiStmt** hStmt, YacHandle* handle);
YapiResult yapiReleaseCursorStmt(YapiStmt* hStmt);

//-----------------------------------------------------------------------------
// Data Function
//-----------------------------------------------------------------------------
YapiResult yapiGetDateStruct(YapiDate date, YapiDateStruct* ds);
YapiResult yapiDateGetDate(const YapiDate date, int16_t* year, uint8_t* month, uint8_t* day);
YapiResult yapiShortTimeGetShortTime(const YapiShortTime time, uint8_t* hour, uint8_t* minute, uint8_t* second,
                                     uint32_t* fraction);
YapiResult yapiTimestampGetTimestamp(const YapiTimestamp timestamp, int16_t* year, uint8_t* month, uint8_t* day,
                                     uint8_t* hour, uint8_t* minute, uint8_t* second, uint32_t* fraction);
YapiResult yapiYMIntervalGetYearMonth(const YapiYMInterval ymInterval, int32_t* year, int32_t* month);
YapiResult yapiDSIntervalGetDaySecond(const YapiDSInterval dsInterval, int32_t* day, int32_t* hour, int32_t* minute,
                                      int32_t* second, int32_t* fraction);

YapiResult yapiDateSetDate(YapiDate* date, int16_t year, uint8_t month, uint8_t day);
YapiResult yapiShortTimeSetShortTime(YapiShortTime* time, uint8_t hour, uint8_t minute, uint8_t second,
                                     uint32_t fraction);
YapiResult yapiTimestampSetTimestamp(YapiTimestamp* timestamp, int16_t year, uint8_t month, uint8_t day, uint8_t hour,
                                     uint8_t minute, uint8_t second, uint32_t fraction);

YapiResult yapiDateTimeGetTimeZoneOffset(YapiEnv* env, YapiTimestamp timestamp, int8_t* hr, int8_t* mm);

YapiResult yapiYMIntervalSetYearMonth(YapiYMInterval* ymInterval, int32_t year, int32_t month);
YapiResult yapiDSIntervalSetDaySecond(YapiDSInterval* dsInterval, int32_t day, int32_t hour, int32_t minute,
                                      int32_t second, int32_t fraction);

YapiResult yapiDSIntervalFromText(YapiEnv* hEnv, YapiDSInterval* dsInterval, const char* str, uint32_t strLen);

YapiResult yapiYMIntervalFromText(YapiEnv* hEnv, YapiYMInterval* ymInterval, const char* str, uint32_t strLen);

YapiResult yapiNumberRound(YapiNumber* n, int32_t precision, int32_t scale);

YapiResult yapiNumberToText(const YapiNumber* number, const char* fmt, uint32_t fmtLength, const char* nlsParam,
                            uint32_t nlsParamLength, char* str, int32_t bufLength, int32_t* length);

YapiResult yapiNumberFromText(const char* str, uint32_t strLength, const char* fmt, uint32_t fmtLength,
                              const char* nlsParam, uint32_t nlsParamLength, YapiNumber* number);

YapiResult yapiNumberFromReal(const YapiPointer rnum, uint32_t length, YapiNumber* number);

YapiResult yapiNumberToReal(const YapiNumber* number, uint32_t length, YapiPointer rsl);

//-----------------------------------------------------------------------------
// Column Desc Function
//-----------------------------------------------------------------------------
uint8_t yapiColumnDescGetPrecision(const YapiColumnDesc* desc);
void    yapiColumnDescSetPrecision(YapiColumnDesc* desc, uint8_t precision);
int8_t  yapiColumnDescGetScale(const YapiColumnDesc* desc);
void    yapiColumnDescSetScale(YapiColumnDesc* desc, int8_t scale);
uint8_t yapiColumnDescGetVectorFormat(const YapiColumnDesc* desc);
void    yapiColumnDescSetVectorFormat(YapiColumnDesc* desc, uint8_t format);

//-----------------------------------------------------------------------------
// Lob Function
//-----------------------------------------------------------------------------
YapiResult yapiLobDescAlloc(YapiConnect* hConn, YapiType type, void** desc);
YapiResult yapiLobDescFree(void* desc, YapiType type);
YapiResult yapiLobGetChunkSize(YapiConnect* hConn, YapiLobLocator* locator, uint16_t* chunkSize);
YapiResult yapiLobGetLength(YapiConnect* hConn, YapiLobLocator* locator, uint64_t* length);
YapiResult yapiLobRead(YapiConnect* hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf, uint64_t bufLen);
YapiResult yapiLobWrite(YapiConnect* hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf, uint64_t bufLen);
YapiResult yapiLobCreateTemporary(YapiConnect* hConn, YapiLobLocator* loc);
YapiResult yapiLobFreeTemporary(YapiConnect* hConn, YapiLobLocator* loc);

//-----------------------------------------------------------------------------
// Vector Function
//-----------------------------------------------------------------------------
YapiResult yapiVectorFromText(YapiVector* vector, YapiVectorFormat format, uint16_t dim, char* text, uint32_t textlen, uint32_t mode);
YapiResult yapiVectorFromArray(YapiVector* vector, YapiVectorFormat format, uint16_t dim, uint8_t* array, uint32_t arrayLen, uint32_t mode);
YapiResult yapiVectorToText(YapiVector* vector, char* text, uint32_t* textlen, uint32_t mode);
YapiResult yapiVectorToArray(YapiVector* vector, YapiVectorFormat format, uint16_t* dim, uint8_t* array, uint32_t* arrayLen, uint32_t mode);
YapiResult yapiVectorGetFormat(YapiVector* vector, YapiVectorFormat* format);
YapiResult yapiVectorGetDimension(YapiVector* vector, uint16_t* dim);

//-----------------------------------------------------------------------------
// plsql debug Function
//-----------------------------------------------------------------------------
YapiResult yapiPdbgStart(YapiStmt* hStmt, uint64_t objId, uint16_t subId);
YapiResult yapiPdbgCheckVersion(YapiStmt* hStmt, uint64_t objId, uint16_t subId, uint32_t version);
YapiResult yapiPdbgAbort(YapiStmt* hStmt);
YapiResult yapiPdbgContinue(YapiStmt* hStmt);
YapiResult yapiPdbgStepInto(YapiStmt* hStmt);
YapiResult yapiPdbgStepOut(YapiStmt* hStmt);
YapiResult yapiPdbgStepNext(YapiStmt* hStmt);

YapiResult yapiPdbgDeleteAllBreakpoints(YapiStmt* hStmt);
YapiResult yapiPdbgAddBreakpoint(YapiStmt* hStmt, uint64_t objId, uint16_t subId, uint32_t lineNo, uint32_t* bpId);
YapiResult yapiPdbgDeleteBreakpoint(YapiStmt* hStmt, uint64_t objId, uint16_t subId, uint32_t lineNo);
YapiResult yapiPdbgGetBreakpointsCount(YapiStmt* hStmt, uint32_t* bpCount);
YapiResult yapiPdbgGetAllVars(YapiStmt* hStmt, uint32_t* bpCount);
YapiResult yapiPdbgGetAllFrames(YapiStmt* hStmt, uint32_t* bpCount);

YapiResult yapiPdbgGetRunningAttrs(YapiStmt* hStmt, YapiDebugRunningAttr attr, YapiPointer value, int32_t bufLen,
                                   int32_t* stringLength);
YapiResult yapiPdbgGetFrameAttrs(YapiStmt* hStmt, uint32_t id, YapiDebugFrameAttr attr, YapiPointer value,
                                 int32_t bufLen, int32_t* stringLength);
YapiResult yapiPdbgGetVarAttrs(YapiStmt* hStmt, uint32_t id, YapiDebugVarAttr attr, YapiPointer value, int32_t bufLen,
                               int32_t* stringLength);
YapiResult yapiPdbgGetVarValue(YapiStmt* hStmt, uint32_t id, uint32_t valueType, YapiPointer value, int32_t bufLen,
                               int32_t* indicator);
YapiResult yapiPdbgGetBreakpointAttrs(YapiStmt* hStmt, uint32_t id, YapiDebugBpAttr attr, YapiPointer value,
                                      int32_t bufLen, int32_t* stringLength);
YapiResult yapiPdbgGetOutput(YapiStmt* hStmt, char* buffer, uint32_t* len, bool* hasMore);

#ifdef __cplusplus
}
#endif

#endif
