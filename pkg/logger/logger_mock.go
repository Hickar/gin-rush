package logger

type loggerMock struct {
	logger
}

func NewLoggerMock() (Logger, error) {
	_logger = &loggerMock{}
	return _logger, nil
}

func (l *loggerMock) Info(message string) {}

func (l *loggerMock) Warning(message string) {}

func (l *loggerMock) Error(err error) {}