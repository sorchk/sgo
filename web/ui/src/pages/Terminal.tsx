import React, { useState, useEffect, useRef } from 'react';
import {
  Box,
  Typography,
  Paper,
  TextField,
  Button,
  Grid,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Send as SendIcon,
} from '@mui/icons-material';
import Layout from '../components/Layout/Layout';
import apiService from '../services/api';

interface TerminalInfo {
  id: string;
  command: string;
  args: string[];
  created_at: string;
}

interface TerminalOutput {
  id: string;
  stdout: string;
  stderr: string;
}

const Terminal: React.FC = () => {
  const [terminals, setTerminals] = useState<TerminalInfo[]>([]);
  const [selectedTerminal, setSelectedTerminal] = useState<string | null>(null);
  const [terminalOutput, setTerminalOutput] = useState<string>('');
  const [command, setCommand] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState<boolean>(false);
  const [newTerminalId, setNewTerminalId] = useState<string>('');
  const [newTerminalCommand, setNewTerminalCommand] = useState<string>('');
  const [actionLoading, setActionLoading] = useState<boolean>(false);
  const outputRef = useRef<HTMLDivElement>(null);
  const pollingRef = useRef<NodeJS.Timeout | null>(null);

  const fetchTerminals = async () => {
    try {
      setLoading(true);
      const response = await apiService.terminals.list();
      setTerminals(response.data.data);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch terminals');
    } finally {
      setLoading(false);
    }
  };

  const fetchTerminalOutput = async (id: string) => {
    try {
      const response = await apiService.terminals.read(id);
      const output: TerminalOutput = response.data.data;
      setTerminalOutput((prev) => {
        // Only append new output if there's something new
        if (output.stdout.trim() || output.stderr.trim()) {
          return prev + output.stdout + output.stderr;
        }
        return prev;
      });
    } catch (err: any) {
      console.error('Failed to read terminal output:', err);
    }
  };

  useEffect(() => {
    fetchTerminals();
    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (selectedTerminal) {
      // Start polling for output
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
      }
      fetchTerminalOutput(selectedTerminal);
      pollingRef.current = setInterval(() => {
        fetchTerminalOutput(selectedTerminal);
      }, 1000);
    } else {
      // Stop polling if no terminal is selected
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    }

    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
      }
    };
  }, [selectedTerminal]);

  useEffect(() => {
    // Scroll to bottom when output changes
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [terminalOutput]);

  const handleSelectTerminal = (id: string) => {
    setSelectedTerminal(id);
    setTerminalOutput('');
  };

  const handleCreateTerminal = async () => {
    if (!newTerminalId) {
      setError('Terminal ID is required');
      return;
    }

    try {
      setActionLoading(true);
      await apiService.terminals.create(
        newTerminalId,
        newTerminalCommand || 'bash',
        []
      );
      setCreateDialogOpen(false);
      setNewTerminalId('');
      setNewTerminalCommand('');
      await fetchTerminals();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create terminal');
    } finally {
      setActionLoading(false);
    }
  };

  const handleKillTerminal = async (id: string) => {
    if (window.confirm(`Are you sure you want to kill terminal ${id}?`)) {
      try {
        setActionLoading(true);
        await apiService.terminals.kill(id);
        if (selectedTerminal === id) {
          setSelectedTerminal(null);
          setTerminalOutput('');
        }
        await fetchTerminals();
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to kill terminal');
      } finally {
        setActionLoading(false);
      }
    }
  };

  const handleSendCommand = async () => {
    if (!selectedTerminal || !command) {
      return;
    }

    try {
      await apiService.terminals.write(selectedTerminal, command + '\n');
      setCommand('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to send command');
    }
  };

  if (loading && terminals.length === 0) {
    return (
      <Layout title="Terminal">
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh">
          <CircularProgress />
        </Box>
      </Layout>
    );
  }

  return (
    <Layout title="Terminal">
      <Typography variant="h4" gutterBottom>
        Terminal Management
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Grid container spacing={2}>
        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2, height: '70vh', display: 'flex', flexDirection: 'column' }}>
            <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
              <Typography variant="h6">Terminals</Typography>
              <Button
                variant="contained"
                size="small"
                startIcon={<AddIcon />}
                onClick={() => setCreateDialogOpen(true)}
              >
                New
              </Button>
            </Box>
            <List sx={{ flexGrow: 1, overflow: 'auto' }}>
              {terminals.map((terminal) => (
                <ListItem
                  key={terminal.id}
                  button
                  selected={selectedTerminal === terminal.id}
                  onClick={() => handleSelectTerminal(terminal.id)}
                >
                  <ListItemText
                    primary={terminal.id}
                    secondary={terminal.command}
                  />
                  <ListItemSecondaryAction>
                    <IconButton
                      edge="end"
                      aria-label="delete"
                      onClick={() => handleKillTerminal(terminal.id)}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </ListItemSecondaryAction>
                </ListItem>
              ))}
              {terminals.length === 0 && (
                <ListItem>
                  <ListItemText primary="No terminals found" />
                </ListItem>
              )}
            </List>
          </Paper>
        </Grid>
        <Grid item xs={12} md={9}>
          <Paper sx={{ p: 2, height: '70vh', display: 'flex', flexDirection: 'column' }}>
            <Typography variant="h6" gutterBottom>
              {selectedTerminal ? `Terminal: ${selectedTerminal}` : 'Select a terminal'}
            </Typography>
            <Box
              ref={outputRef}
              sx={{
                flexGrow: 1,
                bgcolor: 'black',
                color: 'lightgreen',
                p: 1,
                fontFamily: 'monospace',
                overflow: 'auto',
                whiteSpace: 'pre-wrap',
                mb: 2,
              }}
            >
              {terminalOutput || 'No output yet'}
            </Box>
            <Box display="flex">
              <TextField
                fullWidth
                variant="outlined"
                placeholder="Enter command..."
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                disabled={!selectedTerminal}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    handleSendCommand();
                  }
                }}
                sx={{ mr: 1 }}
              />
              <Button
                variant="contained"
                endIcon={<SendIcon />}
                onClick={handleSendCommand}
                disabled={!selectedTerminal || !command}
              >
                Send
              </Button>
            </Box>
          </Paper>
        </Grid>
      </Grid>

      {/* Create Terminal Dialog */}
      <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)}>
        <DialogTitle>Create New Terminal</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Terminal ID"
            fullWidth
            value={newTerminalId}
            onChange={(e) => setNewTerminalId(e.target.value)}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Command (optional)"
            fullWidth
            value={newTerminalCommand}
            onChange={(e) => setNewTerminalCommand(e.target.value)}
            helperText="Default: bash"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleCreateTerminal}
            variant="contained"
            disabled={!newTerminalId || actionLoading}
          >
            {actionLoading ? <CircularProgress size={24} /> : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Layout>
  );
};

export default Terminal;
