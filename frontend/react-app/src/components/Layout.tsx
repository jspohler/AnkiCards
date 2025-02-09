import { ReactNode } from 'react';
import { AppBar, Box, Container, Toolbar, Typography, Button } from '@mui/material';
import { Link, useNavigate } from 'react-router-dom';
import MenuBookIcon from '@mui/icons-material/MenuBook';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';

interface LayoutProps {
  children: ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const navigate = useNavigate();

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar position="static">
        <Toolbar sx={{ display: 'flex', justifyContent: 'space-between' }}>
          <Typography
            variant="h6"
            component={Link}
            to="/"
            sx={{ textDecoration: 'none', color: 'inherit' }}
          >
            AnkiCards Generator
          </Typography>
          
          <Box>
            <Button
              color="inherit"
              startIcon={<CloudUploadIcon />}
              onClick={() => navigate('/')}
              sx={{ mr: 2 }}
            >
              Upload PDF
            </Button>
            <Button
              color="inherit"
              startIcon={<MenuBookIcon />}
              onClick={() => navigate('/decks')}
            >
              View Decks
            </Button>
          </Box>
        </Toolbar>
      </AppBar>
      <Container component="main" sx={{ mt: 4, mb: 4, flex: 1 }}>
        {children}
      </Container>
    </Box>
  );
} 