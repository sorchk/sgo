import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  IconButton,
  Breadcrumbs,
  Link,
  CircularProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from '@mui/material';
import {
  Folder as FolderIcon,
  InsertDriveFile as FileIcon,
  ArrowUpward as UploadIcon,
  CreateNewFolder as NewFolderIcon,
  Delete as DeleteIcon,
  Download as DownloadIcon,
} from '@mui/icons-material';
import Layout from '../components/Layout/Layout';
import apiService from '../services/api';

interface FileInfo {
  path: string;
  size: number;
  mode: string;
  mod_time: string;
  is_dir: boolean;
  md5?: string;
}

const Files: React.FC = () => {
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [currentPath, setCurrentPath] = useState<string>('.');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [uploadOpen, setUploadOpen] = useState<boolean>(false);
  const [newFolderOpen, setNewFolderOpen] = useState<boolean>(false);
  const [newFolderName, setNewFolderName] = useState<string>('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadPath, setUploadPath] = useState<string>('');
  const [actionLoading, setActionLoading] = useState<boolean>(false);

  const fetchFiles = async (path: string = '.') => {
    try {
      setLoading(true);
      const response = await apiService.files.list(path);
      setFiles(response.data.data);
      setCurrentPath(path);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch files');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFiles();
  }, []);

  const handleNavigate = (path: string) => {
    fetchFiles(path);
  };

  const handleNavigateUp = () => {
    if (currentPath === '.' || currentPath === '/') {
      return;
    }
    const parts = currentPath.split('/');
    parts.pop();
    const parentPath = parts.join('/') || '.';
    fetchFiles(parentPath);
  };

  const handleFileClick = (file: FileInfo) => {
    if (file.is_dir) {
      handleNavigate(file.path);
    }
  };

  const handleUploadClick = () => {
    setUploadPath('');
    setSelectedFile(null);
    setUploadOpen(true);
  };

  const handleNewFolderClick = () => {
    setNewFolderName('');
    setNewFolderOpen(true);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) {
      setError('Please select a file to upload');
      return;
    }

    try {
      setActionLoading(true);
      const path = uploadPath || `${currentPath === '.' ? '' : currentPath}/${selectedFile.name}`;
      await apiService.files.upload(selectedFile, path);
      setUploadOpen(false);
      fetchFiles(currentPath);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to upload file');
    } finally {
      setActionLoading(false);
    }
  };

  const handleCreateFolder = async () => {
    if (!newFolderName) {
      setError('Please enter a folder name');
      return;
    }

    try {
      setActionLoading(true);
      const path = `${currentPath === '.' ? '' : currentPath}/${newFolderName}`;
      await apiService.files.mkdir(path);
      setNewFolderOpen(false);
      fetchFiles(currentPath);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create folder');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDownload = (file: FileInfo) => {
    apiService.files.download(file.path);
  };

  const handleDelete = async (file: FileInfo) => {
    if (window.confirm(`Are you sure you want to delete ${file.path}?`)) {
      try {
        setActionLoading(true);
        await apiService.files.delete(file.path);
        fetchFiles(currentPath);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to delete file');
      } finally {
        setActionLoading(false);
      }
    }
  };

  const formatSize = (size: number): string => {
    if (size < 1024) {
      return `${size} B`;
    } else if (size < 1024 * 1024) {
      return `${(size / 1024).toFixed(2)} KB`;
    } else if (size < 1024 * 1024 * 1024) {
      return `${(size / (1024 * 1024)).toFixed(2)} MB`;
    } else {
      return `${(size / (1024 * 1024 * 1024)).toFixed(2)} GB`;
    }
  };

  const formatDate = (dateStr: string): string => {
    const date = new Date(dateStr);
    return date.toLocaleString();
  };

  const renderBreadcrumbs = () => {
    const parts = currentPath === '.' ? [] : currentPath.split('/');
    return (
      <Breadcrumbs aria-label="breadcrumb" sx={{ mb: 2 }}>
        <Link
          component="button"
          underline="hover"
          color="inherit"
          onClick={() => handleNavigate('.')}
        >
          Home
        </Link>
        {parts.map((part, index) => {
          const path = parts.slice(0, index + 1).join('/');
          return (
            <Link
              key={path}
              component="button"
              underline="hover"
              color="inherit"
              onClick={() => handleNavigate(path)}
            >
              {part}
            </Link>
          );
        })}
      </Breadcrumbs>
    );
  };

  if (loading && files.length === 0) {
    return (
      <Layout title="Files">
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh">
          <CircularProgress />
        </Box>
      </Layout>
    );
  }

  return (
    <Layout title="Files">
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
        <Typography variant="h4">File Management</Typography>
        <Box>
          <Button
            variant="contained"
            startIcon={<UploadIcon />}
            onClick={handleUploadClick}
            sx={{ mr: 1 }}
          >
            Upload
          </Button>
          <Button
            variant="outlined"
            startIcon={<NewFolderIcon />}
            onClick={handleNewFolderClick}
          >
            New Folder
          </Button>
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {renderBreadcrumbs()}

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Size</TableCell>
              <TableCell>Modified</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {currentPath !== '.' && (
              <TableRow>
                <TableCell colSpan={4}>
                  <Button startIcon={<FolderIcon />} onClick={handleNavigateUp}>
                    ..
                  </Button>
                </TableCell>
              </TableRow>
            )}
            {files.map((file) => (
              <TableRow key={file.path}>
                <TableCell>
                  <Box display="flex" alignItems="center">
                    {file.is_dir ? (
                      <FolderIcon color="primary" sx={{ mr: 1 }} />
                    ) : (
                      <FileIcon color="action" sx={{ mr: 1 }} />
                    )}
                    <Link
                      component="button"
                      underline="hover"
                      onClick={() => handleFileClick(file)}
                    >
                      {file.path.split('/').pop()}
                    </Link>
                  </Box>
                </TableCell>
                <TableCell>{file.is_dir ? '-' : formatSize(file.size)}</TableCell>
                <TableCell>{formatDate(file.mod_time)}</TableCell>
                <TableCell>
                  <Box display="flex">
                    {!file.is_dir && (
                      <IconButton
                        color="primary"
                        onClick={() => handleDownload(file)}
                        size="small"
                      >
                        <DownloadIcon />
                      </IconButton>
                    )}
                    <IconButton
                      color="error"
                      onClick={() => handleDelete(file)}
                      size="small"
                    >
                      <DeleteIcon />
                    </IconButton>
                  </Box>
                </TableCell>
              </TableRow>
            ))}
            {files.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} align="center">
                  No files found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Upload Dialog */}
      <Dialog open={uploadOpen} onClose={() => setUploadOpen(false)}>
        <DialogTitle>Upload File</DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2 }}>
            <input
              type="file"
              onChange={handleFileSelect}
              style={{ marginBottom: '16px', display: 'block' }}
            />
            <TextField
              label="Upload Path (optional)"
              fullWidth
              value={uploadPath}
              onChange={(e) => setUploadPath(e.target.value)}
              helperText={`If not specified, the file will be uploaded to ${
                currentPath === '.' ? 'the root directory' : currentPath
              }`}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setUploadOpen(false)}>Cancel</Button>
          <Button
            onClick={handleUpload}
            variant="contained"
            disabled={!selectedFile || actionLoading}
          >
            {actionLoading ? <CircularProgress size={24} /> : 'Upload'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* New Folder Dialog */}
      <Dialog open={newFolderOpen} onClose={() => setNewFolderOpen(false)}>
        <DialogTitle>Create New Folder</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Folder Name"
            fullWidth
            value={newFolderName}
            onChange={(e) => setNewFolderName(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setNewFolderOpen(false)}>Cancel</Button>
          <Button
            onClick={handleCreateFolder}
            variant="contained"
            disabled={!newFolderName || actionLoading}
          >
            {actionLoading ? <CircularProgress size={24} /> : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Layout>
  );
};

export default Files;
