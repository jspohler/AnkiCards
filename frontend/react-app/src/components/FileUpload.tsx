import { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { 
  Box, 
  Button, 
  Typography, 
  CircularProgress, 
  FormControlLabel,
  Switch,
  Paper,
  Tooltip
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

export default function FileUpload() {
  const [files, setFiles] = useState<File[]>([]);
  const [uploading, setUploading] = useState(false);
  const [includeTopicCards, setIncludeTopicCards] = useState(true);
  const navigate = useNavigate();

  const onDrop = useCallback((acceptedFiles: File[]) => {
    setFiles(prev => [...prev, ...acceptedFiles]);
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/pdf': ['.pdf']
    }
  });

  const handleUpload = async () => {
    if (files.length === 0) return;

    setUploading(true);
    const formData = new FormData();
    files.forEach(file => {
      formData.append('files', file);
    });

    try {
      const uploadResponse = await axios.post('/api/upload', formData);
      const processResponse = await axios.post('/api/process', {
        files: files.map(f => f.name),
        includeTopicCards: includeTopicCards,
        cardsPerTopic: 5
      });

      navigate(`/status/${processResponse.data.jobId}`);
    } catch (error) {
      console.error('Upload failed:', error);
      // TODO: Add error handling UI
    } finally {
      setUploading(false);
    }
  };

  return (
    <Box sx={{ maxWidth: 600, mx: 'auto' }}>
      <Typography variant="h4" gutterBottom align="center">
        Upload PDF Files
      </Typography>
      
      <Paper sx={{ p: 3, mb: 3 }}>
        <Box
          {...getRootProps()}
          sx={{
            border: '2px dashed',
            borderColor: isDragActive ? 'primary.main' : 'grey.300',
            borderRadius: 2,
            p: 3,
            mb: 2,
            cursor: 'pointer'
          }}
        >
          <input {...getInputProps()} />
          <Typography align="center">
            {isDragActive
              ? 'Drop the files here...'
              : 'Drag and drop PDF files here, or click to select files'}
          </Typography>
        </Box>

        <Tooltip title="Generate additional cards that summarize and connect the main concepts from the document">
          <FormControlLabel
            control={
              <Switch
                checked={includeTopicCards}
                onChange={(e) => setIncludeTopicCards(e.target.checked)}
                color="primary"
              />
            }
            label="Include Topic Summary Cards"
          />
        </Tooltip>
      </Paper>

      {files.length > 0 && (
        <Paper sx={{ p: 3 }}>
          <Typography variant="subtitle1" gutterBottom>
            Selected files:
          </Typography>
          {files.map((file, index) => (
            <Typography key={index} variant="body2">
              {file.name}
            </Typography>
          ))}
          <Button
            variant="contained"
            onClick={handleUpload}
            disabled={uploading}
            fullWidth
            sx={{ mt: 2 }}
          >
            {uploading ? <CircularProgress size={24} /> : 'Upload and Process'}
          </Button>
        </Paper>
      )}
    </Box>
  );
} 