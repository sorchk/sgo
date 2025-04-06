import React, { useState } from 'react';
import {
  Box,
  Typography,
  Paper,
  TextField,
  Button,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  Send as SendIcon,
  Clear as ClearIcon,
} from '@mui/icons-material';
import Layout from '../components/Layout/Layout';
import apiService from '../services/api';

const Shell: React.FC = () => {
  const [command, setCommand] = useState<string>('');
  const [output, setOutput] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleExecuteCommand = async () => {
    if (!command) {
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const response = await apiService.shell.execute(command);
      setOutput((prev) => `${prev}$ ${command}\n${response.data.data.output}\n`);
      setCommand('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to execute command');
    } finally {
      setLoading(false);
    }
  };

  const handleClearOutput = () => {
    setOutput('');
  };

  return (
    <Layout title="Shell">
      <Typography variant="h4" gutterBottom>
        Shell Command Execution
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Paper sx={{ p: 2, mb: 2 }}>
        <Box display="flex" alignItems="center">
          <TextField
            fullWidth
            label="Command"
            variant="outlined"
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            onKeyPress={(e) => {
              if (e.key === 'Enter') {
                handleExecuteCommand();
              }
            }}
            disabled={loading}
            sx={{ mr: 1 }}
          />
          <Button
            variant="contained"
            color="primary"
            startIcon={<SendIcon />}
            onClick={handleExecuteCommand}
            disabled={!command || loading}
          >
            {loading ? <CircularProgress size={24} /> : 'Execute'}
          </Button>
        </Box>
      </Paper>

      <Paper sx={{ p: 2, position: 'relative' }}>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
          <Typography variant="h6">Output</Typography>
          <Button
            variant="outlined"
            size="small"
            startIcon={<ClearIcon />}
            onClick={handleClearOutput}
            disabled={!output}
          >
            Clear
          </Button>
        </Box>
        <Box
          sx={{
            bgcolor: 'black',
            color: 'lightgreen',
            p: 2,
            fontFamily: 'monospace',
            height: '60vh',
            overflow: 'auto',
            whiteSpace: 'pre-wrap',
          }}
        >
          {output || 'No output yet. Execute a command to see results.'}
        </Box>
      </Paper>
    </Layout>
  );
};

export default Shell;
