import React, { useState, useEffect } from 'react';
import {
  Box,
  Container,
  Paper,
  Typography,
  Card,
  CardContent,
  Button,
  List,
  ListItem,
  ListItemText,
  Divider,
} from '@mui/material';
import {
  AccountBalance,
  AccountBalanceWallet,
  Send,
  Receipt,
  History,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import apiService from '../services/api';

interface Balance {
  currency_symbol: string;
  chain_type: string;
  balance: number;
  frozen: number;
  total: number;
}

const Dashboard: React.FC = () => {
  const [balances, setBalances] = useState<Balance[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    loadBalances();
  }, []);

  const loadBalances = async () => {
    try {
      const data = await apiService.getBalances();
      setBalances(data);
    } catch (error) {
      console.error('Failed to load balances:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    navigate('/login');
  };

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Blockchain Wallet Dashboard
        </Typography>
        <Button variant="outlined" onClick={handleLogout}>
          Logout
        </Button>
      </Box>

      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '2fr 1fr' }, gap: 3 }}>
        {/* 余额概览 */}
        <Box>
          <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
            <Typography variant="h6" gutterBottom>
              <AccountBalance sx={{ mr: 1, verticalAlign: 'middle' }} />
              Asset Overview
            </Typography>
            {loading ? (
              <Typography>Loading balances...</Typography>
            ) : (
              <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2 }}>
                {balances.map((balance) => (
                  <Box key={`${balance.currency_symbol}-${balance.chain_type}`}>
                    <Card>
                      <CardContent>
                        <Typography variant="h6" color="primary">
                          {balance.currency_symbol}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          {balance.chain_type.toUpperCase()}
                        </Typography>
                        <Typography variant="h5" sx={{ mt: 1 }}>
                          {balance.total.toFixed(6)}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          Available: {balance.balance.toFixed(6)}
                        </Typography>
                        {balance.frozen > 0 && (
                          <Typography variant="body2" color="text.secondary">
                            Frozen: {balance.frozen.toFixed(6)}
                          </Typography>
                        )}
                      </CardContent>
                    </Card>
                  </Box>
                ))}
              </Box>
            )}
          </Paper>
        </Box>

        {/* 快速操作 */}
        <Box>
          <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
            <Typography variant="h6" gutterBottom>
              Quick Actions
            </Typography>
            <List>
              <ListItem onClick={() => navigate('/addresses')} sx={{ cursor: 'pointer' }}>
                <AccountBalanceWallet sx={{ mr: 2 }} />
                <ListItemText primary="Manage Addresses" />
              </ListItem>
              <Divider />
              <ListItem onClick={() => navigate('/withdraw')} sx={{ cursor: 'pointer' }}>
                <Send sx={{ mr: 2 }} />
                <ListItemText primary="Withdraw" />
              </ListItem>
              <Divider />
              <ListItem onClick={() => navigate('/deposits')} sx={{ cursor: 'pointer' }}>
                <Receipt sx={{ mr: 2 }} />
                <ListItemText primary="Deposit History" />
              </ListItem>
              <Divider />
              <ListItem onClick={() => navigate('/transactions')} sx={{ cursor: 'pointer' }}>
                <History sx={{ mr: 2 }} />
                <ListItemText primary="Transaction History" />
              </ListItem>
            </List>
          </Paper>
        </Box>

        {/* 最近交易 */}
        <Box sx={{ gridColumn: '1 / -1' }}>
          <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
            <Typography variant="h6" gutterBottom>
              Recent Transactions
            </Typography>
            <Typography variant="body2" color="text.secondary">
              No recent transactions
            </Typography>
          </Paper>
        </Box>
      </Box>
    </Container>
  );
};

export default Dashboard; 